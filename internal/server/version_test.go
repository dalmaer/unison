package server

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/digitalex/unison/internal/store"
	"github.com/digitalex/unison/internal/store/storetest"
)

func TestServerVersionHeaderAndEndpoint(t *testing.T) {
	st, err := store.Open(storetest.DSN(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	s := New(st, nil)
	ts := httptest.NewServer(s)
	t.Cleanup(ts.Close)

	// Unstamped dev build: no X-Unison-Version header, /api/version returns "dev".
	resp, err := http.Get(ts.URL + "/api/version")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if got := resp.Header.Get("X-Unison-Version"); got != "" {
		t.Fatalf("unstamped X-Unison-Version = %q, want empty", got)
	}
	var body struct {
		Version string `json:"version"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.Version != "dev" {
		t.Fatalf("unstamped /api/version = %q, want dev", body.Version)
	}

	// Stamped build: sets X-Unison-Version on /api/ routes and returns it on /api/version.
	s.SetVersion("v1.2.3")
	resp2, err := http.Get(ts.URL + "/api/version")
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()
	if got := resp2.Header.Get("X-Unison-Version"); got != "v1.2.3" {
		t.Fatalf("stamped X-Unison-Version = %q, want v1.2.3", got)
	}
	if err := json.NewDecoder(resp2.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.Version != "v1.2.3" {
		t.Fatalf("stamped /api/version = %q, want v1.2.3", body.Version)
	}
}

func TestWebAssetVersioningAndCacheHeaders(t *testing.T) {
	st, err := store.Open(storetest.DSN(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	web := fstest.MapFS{
		"index.html":     {Data: []byte(`<html><script src="/assets/app.js"></script><link href="/assets/app.css"></html>`)},
		"assets/app.js":  {Data: []byte("console.log(1)")},
		"assets/app.css": {Data: []byte("body{}")},
	}
	s := New(st, web)
	if len(s.webVersion) != 12 {
		t.Fatalf("webVersion = %q, want 12 chars", s.webVersion)
	}
	ts := httptest.NewServer(s)
	t.Cleanup(ts.Close)

	// index.html has ?v=<hash> and Cache-Control: no-cache.
	resp, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if got := resp.Header.Get("Cache-Control"); got != "no-cache" {
		t.Fatalf("index Cache-Control = %q, want no-cache", got)
	}
	wantSuffix := "?v=" + s.webVersion
	if !strings.Contains(string(raw), "/assets/app.js"+wantSuffix) || !strings.Contains(string(raw), "/assets/app.css"+wantSuffix) {
		t.Fatalf("index.html not rewritten with %s: %s", wantSuffix, raw)
	}

	// Versioned asset gets immutable cache + ETag.
	r2, err := http.Get(ts.URL + "/assets/app.js" + wantSuffix)
	if err != nil {
		t.Fatal(err)
	}
	r2.Body.Close()
	if got := r2.Header.Get("Cache-Control"); got != "public, max-age=31536000, immutable" {
		t.Fatalf("versioned asset Cache-Control = %q", got)
	}
	if got := r2.Header.Get("ETag"); got != `"`+s.webVersion+`"` {
		t.Fatalf("versioned asset ETag = %q", got)
	}

	// Unversioned or stale asset gets no-cache.
	r3, err := http.Get(ts.URL + "/assets/app.js?v=stale")
	if err != nil {
		t.Fatal(err)
	}
	r3.Body.Close()
	if got := r3.Header.Get("Cache-Control"); got != "no-cache" {
		t.Fatalf("stale asset Cache-Control = %q, want no-cache", got)
	}
}
