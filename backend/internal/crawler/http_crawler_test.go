package crawler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHTTPCrawler_Crawl_HappyPath(t *testing.T) {
	t.Parallel()

	// Test server that returns:
	// - main page with: doctype, title, headings, links, password input
	// - /ok returns 200
	// - /broken returns 404
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// sample HTML page to crawl
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<!DOCTYPE html>
<html>
  <head><title>Hello Sykell</title></head>
  <body>
    <h1>H1</h1>
    <h2>H2</h2>
    <h2>H2 again</h2>

    <form>
      <input type="password" name="pwd" />
    </form>

    <a href="/ok">internal ok</a>
    <a href="/broken">internal broken</a>
    <a href="https://example.com/">external</a>
  </body>
</html>`))
	})

	mux.HandleFunc("/ok", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/broken", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := newHTTPCrawler()
	// speed up link checks a bit for tests
	c.linkCheckTimeout = 2 * time.Second
	c.linkCheckParallel = 4

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	got, err := c.Crawl(ctx, srv.URL+"/")
	if err != nil {
		t.Fatalf("Crawl() unexpected error: %v", err)
	}

	// Title
	if got.Title != "Hello Sykell" {
		t.Fatalf("Title = %q, want %q", got.Title, "Hello Sykell")
	}

	// HTML version (from doctype)
	// Note: depending on your detectHTMLVersion implementation, adjust expectation if needed.
	if !strings.Contains(strings.ToLower(got.HTMLVersion), "html") {
		t.Fatalf("HTMLVersion = %q, expected to contain %q", got.HTMLVersion, "html")
	}

	// Headings
	if got.H1Count != 1 {
		t.Fatalf("H1Count = %d, want %d", got.H1Count, 1)
	}
	if got.H2Count != 2 {
		t.Fatalf("H2Count = %d, want %d", got.H2Count, 2)
	}

	// Login form presence (password input)
	if got.HasLoginForm != true {
		t.Fatalf("HasLoginForm = %v, want %v", got.HasLoginForm, true)
	}

	// Links: /ok + /broken + https://example.com
	if got.LinksCount != 3 {
		t.Fatalf("LinksCount = %d, want %d", got.LinksCount, 3)
	}
	if got.InternalLinksCount != 2 {
		t.Fatalf("InternalLinksCount = %d, want %d", got.InternalLinksCount, 2)
	}
	if got.ExternalLinksCount != 1 {
		t.Fatalf("ExternalLinksCount = %d, want %d", got.ExternalLinksCount, 1)
	}

	// Inaccessible: /broken is 404 => 1
	if got.InaccessibleLinksCount != 1 {
		t.Fatalf("InaccessibleLinksCount = %d, want %d", got.InaccessibleLinksCount, 1)
	}
}
