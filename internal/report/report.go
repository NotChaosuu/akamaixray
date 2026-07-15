package report

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/NotChaosuu/akamaixray/internal/cookies"
	"github.com/NotChaosuu/akamaixray/internal/headers"
	"github.com/NotChaosuu/akamaixray/internal/identify"
)

var (
	useColor = detectColor()
	bold     = ansi("\x1b[1m")
	dim      = ansi("\x1b[2m")
	reset    = ansi("\x1b[0m")
	green    = ansi("\x1b[92m")
	yellow   = ansi("\x1b[93m")
	red      = ansi("\x1b[91m")
	cyan     = ansi("\x1b[96m")
)

func detectColor() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	fi, err := os.Stderr.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

func ansi(code string) string {
	if !useColor {
		return ""
	}
	return code
}

func Pretty(w io.Writer, v *identify.Verdict) {
	fmt.Fprintln(w)
	fmt.Fprintf(w, "  %sakamaixray verdict%s\n", bold, reset)
	fmt.Fprintf(w, "  %starget%s: %s\n", dim, reset, v.URL)
	fmt.Fprintln(w)

	switch {
	case v.BotManager && v.Family == "v3":
		fmt.Fprintf(w, "  %sAkamai Bot Manager v3%s %s(%s)%s\n", red+bold, reset, dim, v.FamilyReason, reset)
	case v.BotManager && v.Family == "v2":
		fmt.Fprintf(w, "  %sAkamai Bot Manager v2%s %s(%s)%s\n", red+bold, reset, dim, v.FamilyReason, reset)
	case v.BotManager:
		fmt.Fprintf(w, "  %sAkamai Bot Manager%s %s(family unknown from a single request)%s\n", red+bold, reset, dim, reset)
	case v.AkamaiEdge:
		fmt.Fprintf(w, "  %sAkamai edge present, Bot Manager not observed on this route%s\n", yellow+bold, reset)
	default:
		fmt.Fprintf(w, "  %sNo Akamai signals observed%s\n", green+bold, reset)
	}
	if v.SBSD {
		fmt.Fprintf(w, "  %s+ SBSD (Server Bot Detect) also active%s\n", yellow, reset)
	}

	if v.ScriptCandidate != "" {
		fmt.Fprintln(w)
		fmt.Fprintf(w, "  %ssensor script candidate%s\n", bold, reset)
		fmt.Fprintf(w, "    %s\n", v.ScriptCandidate)
		if v.ScriptSize > 0 {
			fmt.Fprintf(w, "    %s%d bytes, %s%s\n", dim, v.ScriptSize, v.ScriptSizeBand, reset)
		}
	}

	if bm := onlyBM(v.CookieMarkers); len(bm) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintf(w, "  %sbot manager cookies%s\n", bold, reset)
		for _, c := range bm {
			fmt.Fprintf(w, "    %s%s%s  %s%s%s\n", cyan, c.Name, reset, dim, c.Semantic, reset)
			if c.AbckHint != "" {
				vc := yellow
				switch c.AbckHint {
				case "accepted (~0~ marker present)":
					vc = green
				case "pre-challenge / rejected (~-1~ marker present)":
					vc = red
				}
				fmt.Fprintf(w, "        verdict: %s%s%s\n", vc, c.AbckHint, reset)
			}
		}
	}

	if len(v.HeaderMarkers) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintf(w, "  %sedge markers%s\n", bold, reset)
		sorted := append([]headers.Marker(nil), v.HeaderMarkers...)
		sort.Slice(sorted, func(i, j int) bool { return sorted[i].Header < sorted[j].Header })
		for _, m := range sorted {
			fmt.Fprintf(w, "    %s%s%s: %s\n", cyan, m.Header, reset, truncate(m.Value, 100))
			fmt.Fprintf(w, "        %s%s%s\n", dim, m.Semantic, reset)
		}
	}

	if len(v.Notes) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintf(w, "  %snotes%s\n", bold, reset)
		for _, n := range v.Notes {
			fmt.Fprintf(w, "    %s- %s%s\n", dim, n, reset)
		}
	}
	fmt.Fprintln(w)
}

func JSON(w io.Writer, v *identify.Verdict) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func CookiesOnly(w io.Writer, v *identify.Verdict, asJSON bool) error {
	if asJSON {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(v.CookieMarkers)
	}
	if len(v.CookieMarkers) == 0 {
		fmt.Fprintln(w, "no cookies observed")
		return nil
	}
	for _, c := range v.CookieMarkers {
		fmt.Fprintf(w, "%s [%s] %s\n", c.Name, c.Kind, c.Semantic)
		if c.AbckHint != "" {
			fmt.Fprintf(w, "  verdict: %s\n", c.AbckHint)
		}
	}
	return nil
}

func HeadersOnly(w io.Writer, v *identify.Verdict, asJSON bool) error {
	if asJSON {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(v.HeaderMarkers)
	}
	if len(v.HeaderMarkers) == 0 {
		fmt.Fprintln(w, "no Akamai-relevant headers observed")
		return nil
	}
	for _, m := range v.HeaderMarkers {
		fmt.Fprintf(w, "%s [%s]\n", m.Header, m.Kind)
		fmt.Fprintf(w, "  value: %s\n", truncate(m.Value, 120))
		fmt.Fprintf(w, "  %s\n", m.Semantic)
	}
	return nil
}

func onlyBM(cs []cookies.Classified) []cookies.Classified {
	var out []cookies.Classified
	for _, c := range cs {
		if c.Kind != cookies.KindOther {
			out = append(out, c)
		}
	}
	return out
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-3] + "..."
}
