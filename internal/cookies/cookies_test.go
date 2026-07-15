package cookies

import (
	"net/http"
	"testing"
)

func TestClassifyKnownCookies(t *testing.T) {
	cases := []struct {
		name     string
		value    string
		wantK    Kind
		wantHint string
	}{
		{"_abck", "ABC~-1~xyz", KindBotManagerVerdict, "pre-challenge / rejected (~-1~ marker present)"},
		{"_abck", "ABC~0~xyz", KindBotManagerVerdict, "accepted (~0~ marker present)"},
		{"_abck", "somethingelse", KindBotManagerVerdict, "no coarse verdict marker present"},
		{"bm_sz", "K~params", KindBotManagerChallenge, ""},
		{"ak_bmsc", "opaque", KindBotManagerSession, ""},
		{"bm_sv", "opaque", KindBotManagerSession, ""},
		{"bm_mi", "opaque", KindBotManagerMobile, ""},
		{"sbsd", "opaque", KindSBSD, ""},
		{"sbsd_o", "opaque", KindSBSD, ""},
		{"RT", "some=beacon&state", KindBoomerangRUM, ""},
		{"random_cookie", "value", KindOther, ""},
	}
	for _, c := range cases {
		got := ClassifyByName(c.name, c.value)
		if got.Kind != c.wantK {
			t.Errorf("%s: kind = %q, want %q", c.name, got.Kind, c.wantK)
		}
		if c.wantHint != "" && got.AbckHint != c.wantHint {
			t.Errorf("%s: abckHint = %q, want %q", c.name, got.AbckHint, c.wantHint)
		}
	}
}

func TestClassifyCaseInsensitive(t *testing.T) {
	got := ClassifyByName("_ABCK", "x~0~y")
	if got.Kind != KindBotManagerVerdict {
		t.Errorf("uppercase _abck should classify as verdict, got %q", got.Kind)
	}
}

func TestHasBotManager(t *testing.T) {
	yes := []Classified{{Kind: KindBotManagerVerdict}, {Kind: KindOther}}
	if !HasBotManager(yes) {
		t.Error("expected HasBotManager=true when a BM cookie is present")
	}
	no := []Classified{{Kind: KindOther}, {Kind: KindBoomerangRUM}}
	if HasBotManager(no) {
		t.Error("expected HasBotManager=false when only non-BM cookies are present")
	}
}

func TestClassifyFromHTTPCookies(t *testing.T) {
	cs := []*http.Cookie{
		{Name: "_abck", Value: "AAA~0~BBB"},
		{Name: "bm_sz", Value: "somechallenge"},
		{Name: "unrelated", Value: "x"},
	}
	got := Classify(cs)
	if len(got) != 3 {
		t.Fatalf("expected 3 classified cookies, got %d", len(got))
	}
	if got[0].Kind != KindBotManagerVerdict {
		t.Errorf("first cookie should be verdict, got %q", got[0].Kind)
	}
	if got[0].AbckHint == "" {
		t.Error("expected an abck hint")
	}
}
