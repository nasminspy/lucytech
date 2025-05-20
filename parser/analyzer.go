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

// AnalysisResult holds the data extracted from the analyzed web page.
type AnalysisResult struct {
	HTMLVersion       string         // Detected HTML version (e.g., HTML 5)
	Title             string         // The page title
	Headings          map[string]int // Count of heading tags (H1, H2, etc.)
	InternalLinks     int            // Number of internal links found on the page
	ExternalLinks     int            // Number of external links found on the page
	InaccessibleLinks int            // Number of links that could not be reached (HTTP errors)
	LoginForm         bool           // True if a password input is found (indicating a login form)
}

// httpClient is reused for all HTTP requests with a timeout, facilitating test mocking.
var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

// AnalyzePage function variable allows overriding for testing/mocking.
var AnalyzePage = realAnalyzePage

// realAnalyzePage performs full page analysis: fetching, parsing, and link checking.
func realAnalyzePage(rawURL string) (*AnalysisResult, error) {
	slog.Info("Starting page analysis", "url", rawURL)

	// Ensure URL has a scheme; default to https:// if missing.
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		rawURL = "https://" + rawURL
		slog.Debug("Prepended https:// to URL", "updated_url", rawURL)
	}

	// Validate the URL format and parse components.
	parsedURL, err := url.ParseRequestURI(rawURL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		slog.Error("Invalid URL format", "error", err, "rawURL", rawURL)
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Fetch the page content via HTTP GET.
	resp, err := httpClient.Get(rawURL)
	if err != nil {
		slog.Error("Failed to fetch URL", "error", err, "url", rawURL)
		return nil, fmt.Errorf("unable to reach URL: %w", err)
	}
	defer resp.Body.Close()

	slog.Debug("Fetched URL", "status_code", resp.StatusCode)
	if resp.StatusCode >= 400 {
		slog.Warn("Received HTTP error status from server", "status_code", resp.StatusCode)
		return nil, fmt.Errorf("HTTP error: %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	// Parse the HTML document from response body.
	doc, err := html.Parse(resp.Body)
	if err != nil {
		slog.Error("Failed to parse HTML document", "error", err)
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Initialize result struct with empty headings map.
	result := &AnalysisResult{Headings: make(map[string]int)}
	var links []string

	// Recursive function to walk through the HTML nodes and extract info.
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "title":
				// Extract page title from <title> tag text content.
				if n.FirstChild != nil {
					result.Title = n.FirstChild.Data
				}
			case "input":
				// Detect login form by presence of password input field.
				for _, attr := range n.Attr {
					if attr.Key == "type" && attr.Val == "password" {
						result.LoginForm = true
					}
				}
			case "a":
				// Collect all href attributes from <a> tags as links.
				for _, attr := range n.Attr {
					if attr.Key == "href" {
						links = append(links, attr.Val)
					}
				}
			default:
				// Count heading tags like h1, h2,... h6 (case insensitive).
				if strings.HasPrefix(n.Data, "h") && len(n.Data) == 2 && n.Data[1] >= '1' && n.Data[1] <= '6' {
					result.Headings[strings.ToUpper(n.Data)]++
				}
			}
		}
		// Recursively process child nodes.
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	// Detect HTML version by examining the document's doctype.
	result.HTMLVersion = detectHTMLVersion(doc)

	// Analyze links: count internal/external and check accessibility concurrently.
	countLinks(result, parsedURL, links)

	slog.Info("Page analysis complete",
		"html_version", result.HTMLVersion,
		"title", result.Title,
		"internal_links", result.InternalLinks,
		"external_links", result.ExternalLinks,
		"inaccessible_links", result.InaccessibleLinks,
		"login_form_detected", result.LoginForm)

	return result, nil
}

// detectHTMLVersion examines the document's doctype to guess the HTML version.
func detectHTMLVersion(doc *html.Node) string {
	for c := doc.FirstChild; c != nil; c = c.NextSibling {
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
	// No doctype found, version unknown
	return "Unknown"
}

const maxConcurrentRequests = 10 // Tune this value based on system capacity

// countLinks counts internal vs external links and checks which links are inaccessible.
// It performs concurrent HTTP HEAD requests to verify link accessibility.
func countLinks(result *AnalysisResult, base *url.URL, links []string) {
	seen := make(map[string]bool)                     // Track processed links to avoid duplicates
	var wg sync.WaitGroup                             // WaitGroup to wait for all link checks
	resultCh := make(chan bool, len(links))           // Buffered channel to collect accessibility results
	sem := make(chan struct{}, maxConcurrentRequests) // Semaphore to limit concurrency

	for _, link := range links {
		if link == "" || seen[link] {
			continue // Skip empty or already processed links
		}
		seen[link] = true

		// Parse the link URL relative to base if it's not absolute
		linkURL, err := url.Parse(link)
		if err != nil {
			slog.Warn("Skipping malformed link", "link", link, "error", err)
			continue
		}
		if !linkURL.IsAbs() {
			linkURL = base.ResolveReference(linkURL)
		}

		// Increment internal or external link counts
		if linkURL.Host == base.Host {
			result.InternalLinks++
		} else {
			result.ExternalLinks++
		}

		// Concurrently check link accessibility via HTTP HEAD request
		wg.Add(1)
		go func(link string) {
			defer wg.Done()

			sem <- struct{}{}        // Acquire a semaphore slot
			defer func() { <-sem }() // Release the semaphore slot

			// Create a HEAD request to avoid downloading the whole content
			req, err := http.NewRequest(http.MethodHead, link, nil)
			if err != nil {
				slog.Warn("Failed to create HEAD request", "link", link, "error", err)
				resultCh <- false
				return
			}

			req.Header.Set("User-Agent", "Golang Link Checker")

			resp, err := httpClient.Do(req)
			if err != nil {
				slog.Warn("HEAD request failed", "link", link, "error", err)
				resultCh <- false
				return
			}
			defer resp.Body.Close()

			// Consider HTTP 400+ responses as inaccessible
			if resp.StatusCode >= 400 {
				slog.Warn("Link returned error status", "link", link, "status_code", resp.StatusCode)
				resultCh <- false
				return
			}
			// Link is accessible
			resultCh <- true
		}(linkURL.String())
	}

	// Close the channel after all goroutines finish
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// Count how many links were inaccessible
	for accessible := range resultCh {
		if !accessible {
			result.InaccessibleLinks++
		}
	}
}
