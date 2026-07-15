package identify

import (
	"context"
	"net/url"
	"regexp"
	"strings"

	"github.com/NotChaosuu/akamaixray/internal/cookies"
	"github.com/NotChaosuu/akamaixray/internal/fetch"
	"github.com/NotChaosuu/akamaixray/internal/headers"
)

type Verdict struct {
	URL             string               `json:"url"`
	AkamaiEdge      bool                 `json:"akamai_edge"`
	BotManager      bool                 `json:"bot_manager"`
	SBSD            bool                 `json:"sbsd"`
	Family          string               `json:"family"`
	FamilyReason    string               `json:"family_reason,omitempty"`
	ScriptCandidate string               `json:"script_candidate,omitempty"`
	ScriptSize      int64                `json:"script_size,omitempty"`
	ScriptSizeBand  string               `json:"script_size_band,omitempty"`
	CookieMarkers   []cookies.Classified `json:"cookies"`
	HeaderMarkers   []headers.Marker     `json:"headers"`
	Notes           []string             `json:"notes,omitempty"`
}

var scriptTagRE = regexp.MustCompile(`(?is)<script[^>]+src\s*=\s*["']([^"']+)["']`)

func Detect(ctx context.Context, target string, opts fetch.Options) (*Verdict, error) {
	page, err := fetch.Get(ctx, target, opts)
	if err != nil {
		return nil, err
	}

	v := &Verdict{URL: page.URL}

	v.HeaderMarkers = headers.Scan(page.Headers)
	for _, m := range v.HeaderMarkers {
		if m.Kind == headers.KindEdgeTiming || m.Kind == headers.KindEdgeTransform ||
			m.Kind == headers.KindEdgeID || m.Kind == headers.KindEdgeCache {
			v.AkamaiEdge = true
		}
	}

	v.CookieMarkers = cookies.Classify(page.SetCookies)
	for _, c := range v.CookieMarkers {
		switch c.Kind {
		case cookies.KindBotManagerVerdict, cookies.KindBotManagerChallenge,
			cookies.KindBotManagerSession, cookies.KindBotManagerMobile:
			v.BotManager = true
			v.AkamaiEdge = true
		case cookies.KindSBSD:
			v.SBSD = true
			v.AkamaiEdge = true
		}
	}

	scriptCandidate := ""
	if fetch.LooksLikeHTML(page.Body) {
		scriptCandidate = pickAkamaiScript(page.Body, page.URL)
	}
	if scriptCandidate != "" {
		size, _, err := fetch.GetSize(ctx, scriptCandidate, opts)
		// anything under 200 KB is a vendor script we failed to filter, not a sensor
		if err == nil && size >= 200*1024 {
			v.ScriptCandidate = scriptCandidate
			v.ScriptSize = size
			v.ScriptSizeBand = sizeBand(size)
		}
	}

	v.Family, v.FamilyReason = decideFamily(v)

	if v.BotManager && v.ScriptCandidate == "" {
		v.Notes = append(v.Notes,
			"Bot Manager cookies present but no sensor script in page HTML. Akamai usually lazy-loads the sensor from its challenge flow, so a plain GET often won't include it. Current deployments are most likely v3.")
	}
	if v.AkamaiEdge && !v.BotManager {
		v.Notes = append(v.Notes,
			"Akamai edge present but no Bot Manager cookies on this route. The route may be un-managed, or the challenge only fires on specific paths.")
	}

	return v, nil
}

// scripts whose names give them away as something other than the sensor
var knownNonSensorMarkers = []string{
	"jquery", "react.", "reactdom", "vendor", "gtag", "gtm",
	"analytics", "google", "hotjar", "segment", "mixpanel", "amplitude",
	"criteo", "doubleclick", "facebook", "fbevents", "pinterest",
	"adobedtm", "launch-", "adobe-", "adobescor", "omniture",
	"ruxit", "dynatrace", "newrelic", "boomerang", "boomr.",
	"livechat", "intercom", "zendesk", "drift.",
}

// pickAkamaiScript returns the first same-origin script that isn't obviously a
// third-party include. An empty result is normal: the sensor is often loaded
// later in the challenge flow, not from the first page.
func pickAkamaiScript(body []byte, pageURL string) string {
	base, err := url.Parse(pageURL)
	if err != nil {
		return ""
	}
	for _, m := range scriptTagRE.FindAllSubmatch(body, -1) {
		if len(m) < 2 {
			continue
		}
		src := strings.TrimSpace(string(m[1]))
		if src == "" {
			continue
		}
		u, err := base.Parse(src)
		if err != nil || u.Host != base.Host {
			continue
		}
		if !strings.HasSuffix(strings.ToLower(u.Path), ".js") {
			continue
		}
		if isKnownNonSensor(u.String()) {
			continue
		}
		return u.String()
	}
	return ""
}

func isKnownNonSensor(s string) bool {
	low := strings.ToLower(s)
	for _, marker := range knownNonSensorMarkers {
		if strings.Contains(low, marker) {
			return true
		}
	}
	return false
}

func sizeBand(n int64) string {
	switch {
	case n >= 500*1024 && n <= 560*1024:
		return "500-560 KB (v3-typical band)"
	case n >= 200*1024 && n < 500*1024:
		return "200-500 KB (older v2-typical band)"
	case n < 200*1024:
		return "under 200 KB (too small for a modern sensor script)"
	default:
		return "over 560 KB (outside current v3 typical band, possible new variant)"
	}
}

func decideFamily(v *Verdict) (string, string) {
	if !v.BotManager {
		return "", ""
	}
	switch {
	case v.ScriptSize >= 500*1024 && v.ScriptSize <= 560*1024:
		return "v3", "script size in the 500-560 KB v3-typical band"
	case v.ScriptSize >= 200*1024 && v.ScriptSize < 500*1024:
		return "v2", "script size in the older v2-typical band"
	case v.ScriptSize > 0:
		return "unknown", "script sized but outside any known typical band"
	}
	return "unknown", "no script size available"
}
