package crawler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
)

type httpCrawler struct {
	client            *http.Client
	linkCheckTimeout  time.Duration
	linkCheckParallel int
	userAgent         string
}

// newHTTPCrawler creates a crawler with sane defaults.
func newHTTPCrawler() *httpCrawler {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		// Timeouts / limits
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,
		IdleConnTimeout:     30 * time.Second,
		DisableCompression:  false,
	}

	return &httpCrawler{
		client: &http.Client{
			Timeout:   12 * time.Second,
			Transport: transport,
			// default redirect policy is fine (up to 10)
		},
		linkCheckTimeout:  6 * time.Second,
		linkCheckParallel: 10,
		userAgent:         "sykell-crawler/1.0",
	}
}

func (c *httpCrawler) Crawl(ctx context.Context, rawURL string) (*Result, error) {
	base, err := url.Parse(rawURL)
	if err != nil || base.Scheme == "" || base.Host == "" {
		return nil, fmt.Errorf("invalid url: %q", rawURL)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch page: %w", err)
	}
	defer resp.Body.Close()

	// read body (limit to avoid huge pages)
	const maxBody = 4 << 20 // 4MB
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBody))
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	// parse HTML
	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("parse html: %w", err)
	}

	out := &Result{}
	out.HTMLVersion = detectHTMLVersion(doc)
	out.Title = extractTitle(doc)

	// headings + login + links (collect hrefs)
	var hrefs []string
	var h1, h2, h3, h4, h5, h6 int
	hasPassword := false

	var walk func(n *html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch strings.ToLower(n.Data) {
			case "a":
				if v, ok := getAttr(n, "href"); ok {
					v = strings.TrimSpace(v)
					if v != "" {
						hrefs = append(hrefs, v)
					}
				}
			case "input":
				if v, ok := getAttr(n, "type"); ok && strings.EqualFold(strings.TrimSpace(v), "password") {
					hasPassword = true
				}
			case "h1":
				h1++
			case "h2":
				h2++
			case "h3":
				h3++
			case "h4":
				h4++
			case "h5":
				h5++
			case "h6":
				h6++
			}
		}
		for ch := n.FirstChild; ch != nil; ch = ch.NextSibling {
			walk(ch)
		}
	}
	walk(doc)

	out.HasLoginForm = hasPassword
	out.H1Count = h1
	out.H2Count = h2
	out.H3Count = h3
	out.H4Count = h4
	out.H5Count = h5
	out.H6Count = h6

	// resolve/classify links
	resolved := make([]*url.URL, 0, len(hrefs))
	internal := 0
	external := 0

	for _, href := range hrefs {
		u2, ok := resolveLink(base, href)
		if !ok {
			// ignore things like mailto:, javascript:, fragments-only, etc.
			continue
		}
		resolved = append(resolved, u2)

		if sameHost(base, u2) {
			internal++
		} else {
			external++
		}
	}

	out.LinksCount = len(resolved)
	out.InternalLinksCount = internal
	out.ExternalLinksCount = external

	// check inaccessible links (HEAD/GET) with limited concurrency
	out.InaccessibleLinksCount = c.countInaccessible(ctx, resolved)

	return out, nil
}

func (c *httpCrawler) countInaccessible(ctx context.Context, links []*url.URL) int {
	if len(links) == 0 {
		return 0
	}

	sem := make(chan struct{}, c.linkCheckParallel)
	var wg sync.WaitGroup

	var mu sync.Mutex
	inaccessible := 0

	for _, u := range links {
		u := u
		wg.Add(1)
		go func() {
			defer wg.Done()

			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				return
			}
			defer func() { <-sem }()

			ok := c.isAccessible(ctx, u.String())
			if !ok {
				mu.Lock()
				inaccessible++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()
	return inaccessible
}

func (c *httpCrawler) isAccessible(ctx context.Context, link string) bool {
	// tries HEAD first, then GET if HEAD fails (some servers don't support HEAD).
	ctx, cancel := context.WithTimeout(ctx, c.linkCheckTimeout)
	defer cancel()

	// HEAD
	if code, err := c.doRequest(ctx, http.MethodHead, link); err == nil {
		// treat 2xx/3xx as accessible
		return code >= 200 && code < 400
	}

	// GET fallback
	if code, err := c.doRequest(ctx, http.MethodGet, link); err == nil {
		return code >= 200 && code < 400
	}
	return false
}

func (c *httpCrawler) doRequest(ctx context.Context, method, link string) (int, error) {
	req, err := http.NewRequestWithContext(ctx, method, link, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.client.Do(req)
	if err != nil {
		return 0, err
	}
	// we don't need body; just close quickly
	io.Copy(io.Discard, io.LimitReader(resp.Body, 16<<10))
	resp.Body.Close()

	return resp.StatusCode, nil
}

func resolveLink(base *url.URL, href string) (*url.URL, bool) {
	href = strings.TrimSpace(href)
	if href == "" {
		return nil, false
	}

	low := strings.ToLower(href)
	switch {
	case strings.HasPrefix(low, "#"):
		return nil, false
	case strings.HasPrefix(low, "mailto:"):
		return nil, false
	case strings.HasPrefix(low, "tel:"):
		return nil, false
	case strings.HasPrefix(low, "javascript:"):
		return nil, false
	}

	u, err := url.Parse(href)
	if err != nil {
		return nil, false
	}

	// resolve relative URLs
	u = base.ResolveReference(u)

	// only http/https
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, false
	}
	if u.Host == "" {
		return nil, false
	}

	// drop fragments
	u.Fragment = ""
	return u, true
}

func sameHost(a, b *url.URL) bool {
	// compare hostname for internal/external classification; ignore port, case-insensitive
	return strings.EqualFold(a.Hostname(), b.Hostname())
}

func extractTitle(doc *html.Node) string {
	var title string
	var walk func(n *html.Node)
	walk = func(n *html.Node) {
		if title != "" {
			return
		}
		if n.Type == html.ElementNode && strings.EqualFold(n.Data, "title") {
			if n.FirstChild != nil && n.FirstChild.Type == html.TextNode {
				title = strings.TrimSpace(n.FirstChild.Data)
				return
			}
		}
		for ch := n.FirstChild; ch != nil; ch = ch.NextSibling {
			walk(ch)
		}
	}
	walk(doc)
	return title
}

func detectHTMLVersion(doc *html.Node) string {
	// tries to classify based on doctype.
	var dt string

	var walk func(n *html.Node)
	walk = func(n *html.Node) {
		if dt != "" {
			return
		}
		if n.Type == html.DoctypeNode {
			// n.Data usually "html"
			// attributes may hold PUBLIC/SYSTEM identifiers
			// We'll reconstruct a string-ish doctype
			sb := strings.Builder{}
			sb.WriteString("<!DOCTYPE ")
			sb.WriteString(n.Data)
			for _, a := range n.Attr {
				// not always populated; depends on parser
				if a.Key != "" && a.Val != "" {
					sb.WriteString(" ")
					sb.WriteString(a.Key)
					sb.WriteString(`="`)
					sb.WriteString(a.Val)
					sb.WriteString(`"`)
				}
			}
			sb.WriteString(">")
			dt = sb.String()
			return
		}
		for ch := n.FirstChild; ch != nil; ch = ch.NextSibling {
			walk(ch)
		}
	}
	walk(doc)

	if dt == "" {
		return "unknown"
	}

	low := strings.ToLower(dt)

	// HTML5: <!doctype html>
	if strings.Contains(low, "<!doctype html") && !strings.Contains(low, "public") && !strings.Contains(low, "system") {
		return "HTML5"
	}

	// XHTML (common)
	if strings.Contains(low, "xhtml") {
		return "XHTML"
	}

	// HTML 4.01
	if strings.Contains(low, "html 4.01") || strings.Contains(low, "html4") {
		return "HTML 4.01"
	}

	// fallback: raw doctype string, trimmed
	return strings.TrimSpace(dt)
}

func getAttr(n *html.Node, key string) (string, bool) {
	for _, a := range n.Attr {
		if strings.EqualFold(a.Key, key) {
			return a.Val, true
		}
	}
	return "", false
}
