package headers

import (
	"net/http"
	"testing"
)

func TestScanClassifiesAkamaiHeaders(t *testing.T) {
	h := http.Header{}
	h.Set("X-Akamai-Transformed", "9 42 0 pmb=mRUM,1")
	h.Set("Server-Timing", `cdn-cache; desc=HIT, ak_p; desc="1234_567"`)
	h.Set("X-Cache", "TCP_HIT")
	h.Set("Content-Type", "text/html")
	h.Add("X-Akamai-Bot-Score", "42") // hypothetical BM marker

	got := Scan(h)
	if len(got) < 3 {
		t.Fatalf("expected at least 3 markers, got %d: %+v", len(got), got)
	}

	sawTransform, sawTiming, sawCache, sawBot := false, false, false, false
	for _, m := range got {
		switch m.Kind {
		case KindEdgeTransform:
			sawTransform = true
		case KindEdgeTiming:
			sawTiming = true
		case KindEdgeCache:
			sawCache = true
		case KindBotManager:
			sawBot = true
		}
	}
	if !sawTransform {
		t.Error("expected x-akamai-transformed classified as EdgeTransform")
	}
	if !sawTiming {
		t.Error("expected server-timing with ak_p classified as EdgeTiming")
	}
	if !sawCache {
		t.Error("expected x-cache classified as EdgeCache")
	}
	if !sawBot {
		t.Error("expected x-akamai-bot-* classified as BotManager")
	}
}

func TestScanIgnoresNonAkamaiHeaders(t *testing.T) {
	h := http.Header{}
	h.Set("Content-Type", "text/html")
	h.Set("Content-Length", "42")
	h.Set("Server", "nginx/1.20.1")

	got := Scan(h)
	if len(got) != 0 {
		t.Errorf("expected 0 markers on plain nginx, got %d: %+v", len(got), got)
	}
}

func TestScanClassifiesAkamaiServer(t *testing.T) {
	h := http.Header{}
	h.Set("Server", "AkamaiGHost")
	got := Scan(h)
	if len(got) != 1 || got[0].Kind != KindEdgeID {
		t.Fatalf("expected 1 EdgeID marker for AkamaiGHost server header, got %+v", got)
	}
}

func TestServerTimingWithoutAkamaiMarker(t *testing.T) {
	h := http.Header{}
	h.Set("Server-Timing", "app; dur=42, db; dur=17")
	got := Scan(h)
	for _, m := range got {
		if m.Kind == KindEdgeTiming {
			t.Errorf("plain server-timing without ak_p should not classify as EdgeTiming: %+v", m)
		}
	}
}
