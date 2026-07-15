package headers

import (
	"net/http"
	"strings"
)

type Kind string

const (
	KindEdgeCache     Kind = "edge_cache"
	KindEdgeTransform Kind = "edge_transform"
	KindEdgeID        Kind = "edge_id"
	KindEdgeTiming    Kind = "edge_timing"
	KindBotManager    Kind = "bot_manager"
	KindOther         Kind = "other"
)

type Marker struct {
	Header   string `json:"header"`
	Value    string `json:"value"`
	Kind     Kind   `json:"kind"`
	Semantic string `json:"semantic,omitempty"`
}

func Scan(h http.Header) []Marker {
	var out []Marker
	for name, vals := range h {
		lname := strings.ToLower(name)
		for _, v := range vals {
			if m, ok := classify(lname, v); ok {
				out = append(out, m)
			}
		}
	}
	return out
}

func classify(lname, v string) (Marker, bool) {
	switch {
	case lname == "x-akamai-transformed":
		return Marker{
			Header:   lname,
			Value:    v,
			Kind:     KindEdgeTransform,
			Semantic: "Response body was transformed by an Akamai edge policy",
		}, true
	case lname == "x-akamai-request-id":
		return Marker{
			Header:   lname,
			Value:    v,
			Kind:     KindEdgeID,
			Semantic: "Akamai edge request correlation ID",
		}, true
	case strings.HasPrefix(lname, "x-akamai-bot"), strings.HasPrefix(lname, "x-akamai-mgmt"):
		return Marker{
			Header:   lname,
			Value:    v,
			Kind:     KindBotManager,
			Semantic: "Bot Manager / management-layer marker (not universally emitted)",
		}, true
	case strings.HasPrefix(lname, "x-akamai-"):
		return Marker{
			Header:   lname,
			Value:    v,
			Kind:     KindEdgeID,
			Semantic: "Akamai-namespaced edge header",
		}, true
	case lname == "server-timing":
		if hasAkamaiTimingMarker(v) {
			return Marker{
				Header:   lname,
				Value:    v,
				Kind:     KindEdgeTiming,
				Semantic: "server-timing with Akamai edge markers (cdn-cache / ak_p / edge / origin)",
			}, true
		}
	case lname == "x-cache", lname == "x-cache-key", lname == "x-check-cacheable":
		return Marker{
			Header:   lname,
			Value:    v,
			Kind:     KindEdgeCache,
			Semantic: "Edge cache diagnostic (commonly Akamai)",
		}, true
	case lname == "x-served-by":
		if looksLikeAkamai(v) {
			return Marker{
				Header:   lname,
				Value:    v,
				Kind:     KindEdgeID,
				Semantic: "Served-by tag consistent with an Akamai edge node",
			}, true
		}
	case lname == "server":
		if looksLikeAkamai(v) {
			return Marker{
				Header:   lname,
				Value:    v,
				Kind:     KindEdgeID,
				Semantic: "Server header names Akamai (AkamaiGHost / AkamaiNetStorage)",
			}, true
		}
	}
	return Marker{}, false
}

func hasAkamaiTimingMarker(v string) bool {
	lower := strings.ToLower(v)
	for _, marker := range []string{"ak_p", "cdn-cache", "edge;", "origin;"} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}

func looksLikeAkamai(v string) bool {
	return strings.Contains(strings.ToLower(v), "akamai")
}

func HasAny(markers []Marker) bool {
	return len(markers) > 0
}
