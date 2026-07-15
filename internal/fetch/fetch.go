package fetch

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Result struct {
	URL        string
	Status     int
	Headers    http.Header
	SetCookies []*http.Cookie
	Body       []byte
	Truncated  bool
	Elapsed    time.Duration
}

type Options struct {
	UserAgent    string
	Timeout      time.Duration
	MaxBytes     int64
	Insecure     bool
	Method       string
	Body         []byte
	ExtraHeaders map[string]string
}

func DefaultOptions() Options {
	return Options{
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/149.0.0.0 Safari/537.36",
		Timeout:   15 * time.Second,
		MaxBytes:  2 * 1024 * 1024,
	}
}

func Get(ctx context.Context, target string, opts Options) (*Result, error) {
	u, err := url.Parse(target)
	if err != nil {
		return nil, fmt.Errorf("invalid url: %w", err)
	}
	if u.Scheme == "" {
		u.Scheme = "https"
	}

	client := &http.Client{
		Timeout: opts.Timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	method := opts.Method
	if method == "" {
		method = "GET"
	}
	var reqBody io.Reader
	if len(opts.Body) > 0 {
		reqBody = bytes.NewReader(opts.Body)
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), reqBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", opts.UserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	// identity encoding keeps body-size and substring checks straightforward
	req.Header.Set("Accept-Encoding", "identity")
	for k, val := range opts.ExtraHeaders {
		req.Header.Set(k, val)
	}

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	limit := opts.MaxBytes
	if limit <= 0 {
		limit = 2 * 1024 * 1024
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, err
	}
	truncated := false
	if int64(len(body)) > limit {
		body = body[:limit]
		truncated = true
	}

	return &Result{
		URL:        u.String(),
		Status:     resp.StatusCode,
		Headers:    resp.Header,
		SetCookies: resp.Cookies(),
		Body:       body,
		Truncated:  truncated,
		Elapsed:    time.Since(start),
	}, nil
}

// GetSize returns the byte size of a URL's body without holding it in memory.
func GetSize(ctx context.Context, target string, opts Options) (int64, string, error) {
	u, err := url.Parse(target)
	if err != nil {
		return 0, "", err
	}
	if u.Scheme == "" {
		u.Scheme = "https"
	}

	client := &http.Client{Timeout: opts.Timeout}
	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return 0, "", err
	}
	req.Header.Set("User-Agent", opts.UserAgent)
	req.Header.Set("Accept-Encoding", "identity")

	resp, err := client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()

	if cl := resp.Header.Get("Content-Length"); cl != "" {
		var n int64
		if _, err := fmt.Sscanf(cl, "%d", &n); err == nil {
			io.Copy(io.Discard, io.LimitReader(resp.Body, 32*1024))
			return n, resp.Header.Get("Content-Type"), nil
		}
	}
	n, err := io.Copy(io.Discard, resp.Body)
	if err != nil {
		return 0, "", err
	}
	return n, resp.Header.Get("Content-Type"), nil
}

func LooksLikeHTML(body []byte) bool {
	if len(body) == 0 {
		return false
	}
	lower := strings.ToLower(string(body[:min(len(body), 512)]))
	return strings.Contains(lower, "<html") ||
		strings.Contains(lower, "<!doctype html") ||
		strings.Contains(lower, "<head") ||
		strings.Contains(lower, "<body")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
