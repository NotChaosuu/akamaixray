package identify

import (
	"testing"
)

func TestSizeBand(t *testing.T) {
	cases := []struct {
		n    int64
		want string
	}{
		{540 * 1024, "500-560 KB (v3-typical band)"},
		{400 * 1024, "200-500 KB (older v2-typical band)"},
		{50 * 1024, "under 200 KB (too small for a modern sensor script)"},
		{600 * 1024, "over 560 KB (outside current v3 typical band, possible new variant)"},
	}
	for _, c := range cases {
		got := sizeBand(c.n)
		if got != c.want {
			t.Errorf("sizeBand(%d) = %q, want %q", c.n, got, c.want)
		}
	}
}

func TestDecideFamilyByScriptSize(t *testing.T) {
	cases := []struct {
		bm    bool
		size  int64
		wantF string
	}{
		{true, 540 * 1024, "v3"},
		{true, 400 * 1024, "v2"},
		{true, 100 * 1024, "unknown"},
		{true, 0, "unknown"},
		{false, 0, ""},
	}
	for _, c := range cases {
		v := &Verdict{BotManager: c.bm, ScriptSize: c.size}
		gotF, _ := decideFamily(v)
		if gotF != c.wantF {
			t.Errorf("BM=%v size=%d -> family=%q, want %q", c.bm, c.size, gotF, c.wantF)
		}
	}
}

func TestPickAkamaiScriptPrefersSameOrigin(t *testing.T) {
	body := []byte(`
<html>
<head>
  <script src="https://www.google-analytics.com/analytics.js"></script>
  <script src="/assets/vendor/jquery.min.js"></script>
  <script src="/xtsyc/8Cd_hHR/vFI7/2vHVJp/DiwmITMi/QG5oIVsl/D2NpV1Zg/BUcSK1EN/UyE/pkg.js"></script>
</head>
</html>`)
	got := pickAkamaiScript(body, "https://target.example.com/")
	want := "https://target.example.com/xtsyc/8Cd_hHR/vFI7/2vHVJp/DiwmITMi/QG5oIVsl/D2NpV1Zg/BUcSK1EN/UyE/pkg.js"
	if got != want {
		t.Errorf("pickAkamaiScript picked %q, want %q", got, want)
	}
}

func TestPickAkamaiScriptSkipsVendorNames(t *testing.T) {
	body := []byte(`
<script src="/vendor/react.js"></script>
<script src="/gtag/js"></script>
<script src="/googleanalytics/ga.js"></script>
<script src="/anonymised/plain.js"></script>`)
	got := pickAkamaiScript(body, "https://target.example.com/")
	want := "https://target.example.com/anonymised/plain.js"
	if got != want {
		t.Errorf("pickAkamaiScript picked %q, want %q", got, want)
	}
}

func TestPickAkamaiScriptEmptyBody(t *testing.T) {
	got := pickAkamaiScript([]byte(""), "https://target.example.com/")
	if got != "" {
		t.Errorf("empty body should return empty candidate, got %q", got)
	}
}
