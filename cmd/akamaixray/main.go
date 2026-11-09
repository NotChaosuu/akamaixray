package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/NotChaosuu/akamaixray/internal/fetch"
	"github.com/NotChaosuu/akamaixray/internal/identify"
	"github.com/NotChaosuu/akamaixray/internal/report"
)

const version = "0.1.1"

// headerList collects repeated --header flags.
type headerList []string

func (h *headerList) String() string { return strings.Join(*h, ", ") }
func (h *headerList) Set(v string) error {
	*h = append(*h, v)
	return nil
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	cmd := os.Args[1]
	switch cmd {
	case "-h", "--help", "help":
		usage()
		return
	case "-v", "--version", "version":
		fmt.Printf("akamaixray %s\n", version)
		return
	case "identify", "cookies", "headers", "scan":
		runSubcommand(cmd, os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", cmd)
		usage()
		os.Exit(2)
	}
}

func runSubcommand(cmd string, args []string) {
	fs := flag.NewFlagSet(cmd, flag.ExitOnError)
	asJSON := fs.Bool("json", false, "emit JSON")
	timeout := fs.Int("timeout", 15, "request timeout in seconds")
	ua := fs.String("ua", "", "override User-Agent")
	method := fs.String("method", "GET", "HTTP method (GET, POST, ...)")
	body := fs.String("body", "", "request body, or @file to read the body from a file")
	var hdrs headerList
	fs.Var(&hdrs, "header", "extra request header 'Name: Value' (repeatable)")
	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}
	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "missing <url>")
		os.Exit(2)
	}
	target := fs.Arg(0)

	opts := fetch.DefaultOptions()
	opts.Timeout = time.Duration(*timeout) * time.Second
	if *ua != "" {
		opts.UserAgent = *ua
	}
	opts.Method = strings.ToUpper(*method)
	if *body != "" {
		b, err := resolveBody(*body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(2)
		}
		opts.Body = b
	}
	if len(hdrs) > 0 {
		opts.ExtraHeaders = parseHeaders(hdrs)
	}

	ctx, cancel := context.WithTimeout(context.Background(), opts.Timeout+5*time.Second)
	defer cancel()

	v, err := identify.Detect(ctx, target, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	switch cmd {
	case "identify", "scan":
		if *asJSON {
			_ = report.JSON(os.Stdout, v)
		} else {
			report.Pretty(os.Stderr, v)
		}
	case "cookies":
		if err := report.CookiesOnly(os.Stdout, v, *asJSON); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "headers":
		if err := report.HeadersOnly(os.Stdout, v, *asJSON); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}

func resolveBody(v string) ([]byte, error) {
	if strings.HasPrefix(v, "@") {
		return os.ReadFile(v[1:])
	}
	return []byte(v), nil
}

func parseHeaders(hdrs headerList) map[string]string {
	m := make(map[string]string, len(hdrs))
	for _, h := range hdrs {
		name, val, ok := strings.Cut(h, ":")
		if !ok {
			fmt.Fprintf(os.Stderr, "warning: ignoring malformed header %q (want 'Name: Value')\n", h)
			continue
		}
		m[strings.TrimSpace(name)] = strings.TrimSpace(val)
	}
	return m
}

func usage() {
	fmt.Fprintf(os.Stderr, `akamaixray %s - passive identification of Akamai Bot Manager on a target

usage:
  akamaixray identify <url>       verdict + family (v2/v3) + script hints
  akamaixray cookies  <url>       semantic labels for observed BM cookies
  akamaixray headers  <url>       Akamai-relevant response headers
  akamaixray scan     <url>       everything (same as identify)

flags (any subcommand):
  --json                    emit JSON to stdout
  --timeout N               request timeout in seconds (default 15)
  --ua "..."                override User-Agent
  --method M                HTTP method (default GET; use POST for login APIs)
  --body "..."              request body, or @file to read from a file
  --header "Name: Value"    extra request header (repeatable)

Reads what the target returns on a single request. On many tenants the
Bot Manager cookies only appear on the protected route (a login/auth POST),
so point --method POST at that endpoint when a homepage GET looks clean.

Docs: github.com/NotChaosuu/akamaixray
`, version)
}
