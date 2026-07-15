package fetch

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetDefaultsToGET(t *testing.T) {
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Method
	}))
	defer srv.Close()

	if _, err := Get(context.Background(), srv.URL, DefaultOptions()); err != nil {
		t.Fatal(err)
	}
	if got != "GET" {
		t.Errorf("method = %q, want GET", got)
	}
}

func TestGetSendsMethodBodyAndHeaders(t *testing.T) {
	var gotMethod, gotBody, gotCT string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		gotCT = r.Header.Get("Content-Type")
		http.SetCookie(w, &http.Cookie{Name: "_abck", Value: "AAA~-1~BBB"})
	}))
	defer srv.Close()

	opts := DefaultOptions()
	opts.Method = "POST"
	opts.Body = []byte(`{"email":"a@b.c","password":"x"}`)
	opts.ExtraHeaders = map[string]string{"Content-Type": "application/json"}

	res, err := Get(context.Background(), srv.URL, opts)
	if err != nil {
		t.Fatal(err)
	}
	if gotMethod != "POST" {
		t.Errorf("method = %q, want POST", gotMethod)
	}
	if gotBody != `{"email":"a@b.c","password":"x"}` {
		t.Errorf("body = %q", gotBody)
	}
	if gotCT != "application/json" {
		t.Errorf("content-type = %q, want application/json", gotCT)
	}
	if len(res.SetCookies) != 1 || res.SetCookies[0].Name != "_abck" {
		t.Errorf("expected _abck cookie back, got %+v", res.SetCookies)
	}
}

func TestExtraHeaderOverridesDefault(t *testing.T) {
	var gotUA string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
	}))
	defer srv.Close()

	opts := DefaultOptions()
	opts.ExtraHeaders = map[string]string{"User-Agent": "custom/1.0"}
	if _, err := Get(context.Background(), srv.URL, opts); err != nil {
		t.Fatal(err)
	}
	if gotUA != "custom/1.0" {
		t.Errorf("user-agent = %q, want custom/1.0 (extra header should win)", gotUA)
	}
}
