package parser

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
)

type AnalysisResult struct {
	HTMLVersion       string
	Title             string
	Headings          map[string]int
	InternalLinks     int
	ExternalLinks     int
	InaccessibleLinks int
	LoginForm         bool
}

func AnalyzePage(rawURL string) (*AnalysisResult, error) {
	slog.Info("Analyzing page", "url", rawURL)

	// Prepend scheme if missing
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		rawURL = "https://" + rawURL
		slog.Debug("Prepended https:// to URL", "updated_url", rawURL)
	}

	// Validate URL structure
	parsedURL, err := url.ParseRequestURI(rawURL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		slog.Error("Invalid URL format", "error", err, "rawURL", rawURL)
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	resp, err := http.Get(rawURL)
	if err != nil {
		slog.Error("Failed to fetch URL", "error", err)
		return nil, fmt.Errorf("unable to reach URL: %w", err)
	}
	defer resp.Body.Close()

	slog.Debug("Fetched URL", "status_code", resp.StatusCode)
	if resp.StatusCode >= 400 {
		slog.Warn("HTTP error from server", "status_code", resp.StatusCode)
		return nil, fmt.Errorf("HTTP error: %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		slog.Error("Failed to parse HTML", "error", err)
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	result := &AnalysisResult{
		Headings: make(map[string]int),
	}
	var f func(*html.Node)
	var links []string

	f = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "title":
				if n.FirstChild != nil {
					result.Title = n.FirstChild.Data
				}
			case "input":
				for _, attr := range n.Attr {
					if attr.Key == "type" && attr.Val == "password" {
						result.LoginForm = true
					}
				}
			case "a":
				for _, attr := range n.Attr {
					if attr.Key == "href" {
						links = append(links, attr.Val)
					}
				}
			default:
				if strings.HasPrefix(n.Data, "h") && len(n.Data) == 2 && n.Data[1] >= '1' && n.Data[1] <= '6' {
					result.Headings[strings.ToUpper(n.Data)]++
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}

	f(doc)
	result.HTMLVersion = detectHTMLVersion(doc)

	countLinks(result, parsedURL, links)

	slog.Info("Page analysis complete",
		"html_version", result.HTMLVersion,
		"title", result.Title,
		"internal_links", result.InternalLinks,
		"external_links", result.ExternalLinks,
		"inaccessible_links", result.InaccessibleLinks,
		"login_form_detected", result.LoginForm,
	)

	return result, nil
}

func detectHTMLVersion(n *html.Node) string {
	for c := n; c != nil; c = c.NextSibling {
		if c.Type == html.DoctypeNode {
			doctype := strings.ToLower(c.Data)
			switch {
			case strings.Contains(doctype, "html 4.01"):
				return "HTML 4.01"
			case strings.Contains(doctype, "xhtml"):
				return "XHTML"
			case strings.Contains(doctype, "html"):
				return "HTML 5"
			default:
				return "Unknown"
			}
		}
	}
	return "Unknown"
}

// countLinks determines internal, external and inaccessible link counts
func countLinks(result *AnalysisResult, base *url.URL, links []string) {
	seen := make(map[string]bool)
	var wg sync.WaitGroup
	resultCh := make(chan bool, len(links)) // buffered to prevent goroutine leaks

	client := http.Client{
		Timeout: 5 * time.Second,
	}

	for _, link := range links {
		if link == "" || seen[link] {
			continue
		}
		seen[link] = true

		linkURL, err := url.Parse(link)
		if err != nil {
			continue
		}

		if !linkURL.IsAbs() {
			linkURL = base.ResolveReference(linkURL)
		}

		// Determine internal or external
		if linkURL.Host == base.Host {
			result.InternalLinks++
		} else {
			result.ExternalLinks++
		}

		// Check accessibility in a goroutine
		wg.Add(1)
		go func(link string) {
			defer wg.Done()
			req, err := http.NewRequest(http.MethodHead, link, nil)
			if err != nil {
				resultCh <- false
				return
			}
			req.Header.Set("User-Agent", "Golang Link Checker")
			resp, err := client.Do(req)
			if err != nil || resp.StatusCode >= 400 {
				resultCh <- false
				return
			}
			resultCh <- true
		}(linkURL.String())
	}

	// Wait for all checks to finish
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// Collect results
	for accessible := range resultCh {
		if !accessible {
			result.InaccessibleLinks++
		}
	}
}
