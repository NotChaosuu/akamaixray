package cookies

import (
	"net/http"
	"strings"
)

type Kind string

const (
	KindBotManagerVerdict   Kind = "bm_verdict"
	KindBotManagerChallenge Kind = "bm_challenge"
	KindBotManagerSession   Kind = "bm_session"
	KindBotManagerMobile    Kind = "bm_mobile"
	KindSBSD                Kind = "sbsd"
	KindBoomerangRUM        Kind = "rum"
	KindOther               Kind = "other"
)

type Classified struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	Kind     Kind   `json:"kind"`
	Product  string `json:"product,omitempty"`
	Semantic string `json:"semantic,omitempty"`
	AbckHint string `json:"abck_hint,omitempty"`
}

func Classify(cs []*http.Cookie) []Classified {
	out := make([]Classified, 0, len(cs))
	for _, c := range cs {
		out = append(out, classifyOne(c.Name, c.Value))
	}
	return out
}

func ClassifyByName(name, value string) Classified {
	return classifyOne(name, value)
}

func classifyOne(name, value string) Classified {
	c := Classified{Name: name, Value: value, Kind: KindOther}
	switch strings.ToLower(name) {
	case "_abck":
		c.Kind = KindBotManagerVerdict
		c.Product = "Akamai Bot Manager"
		c.Semantic = "Mitigation-state verdict token"
		c.AbckHint = abckHint(value)
	case "bm_sz":
		c.Kind = KindBotManagerChallenge
		c.Product = "Akamai Bot Manager"
		c.Semantic = "Challenge-parameter bundle read by the client script"
	case "ak_bmsc":
		c.Kind = KindBotManagerSession
		c.Product = "Akamai Bot Manager"
		c.Semantic = "Session mitigation-state co-cookie"
	case "bm_sv":
		c.Kind = KindBotManagerSession
		c.Product = "Akamai Bot Manager"
		c.Semantic = "Session-verification token"
	case "bm_mi":
		c.Kind = KindBotManagerMobile
		c.Product = "Akamai Bot Manager (mobile SDK path)"
		c.Semantic = "Mobile-integration marker"
	case "bm_so":
		c.Kind = KindBotManagerSession
		c.Product = "Akamai Bot Manager"
		c.Semantic = "Short-lived session-origin binding"
	case "bm_lso":
		c.Kind = KindBotManagerSession
		c.Product = "Akamai Bot Manager"
		c.Semantic = "Long-session origin binding"
	case "sbsd":
		c.Kind = KindSBSD
		c.Product = "Akamai SBSD"
		c.Semantic = "Server Bot Detect short-lived challenge cookie"
	case "sbsd_o":
		c.Kind = KindSBSD
		c.Product = "Akamai SBSD"
		c.Semantic = "SBSD long-form state"
	case "rt":
		c.Kind = KindBoomerangRUM
		c.Product = "Akamai Boomerang (RUM)"
		c.Semantic = "Real-user monitoring beacon state (not a bot cookie, but a signal source)"
	}
	return c
}

func abckHint(v string) string {
	switch {
	case strings.Contains(v, "~0~"):
		return "accepted (~0~ marker present)"
	case strings.Contains(v, "~-1~"):
		return "pre-challenge / rejected (~-1~ marker present)"
	default:
		return "no coarse verdict marker present"
	}
}

func HasBotManager(cs []Classified) bool {
	for _, c := range cs {
		switch c.Kind {
		case KindBotManagerVerdict, KindBotManagerChallenge, KindBotManagerSession,
			KindBotManagerMobile, KindSBSD:
			return true
		}
	}
	return false
}
