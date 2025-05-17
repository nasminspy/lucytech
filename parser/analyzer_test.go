package parser

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

// mockRoundTripper mocks the behavior of http.Client.Transport to simulate HTTP responses
type mockRoundTripper struct {
	// mockGet handles HTTP GET requests
	mockGet func(req *http.Request) *http.Response
	// mockHead handles HTTP HEAD requests
	mockHead func(req *http.Request) *http.Response
}

// RoundTrip implements the RoundTripper interface for mockRoundTripper
func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// For GET requests, call the mockGet function
	if req.Method == http.MethodGet {
		return m.mockGet(req), nil
	} else if req.Method == http.MethodHead {
		// For HEAD requests, call the mockHead function
		return m.mockHead(req), nil
	}
	// Default fallback for any other HTTP methods returns 404 with empty body
	return &http.Response{
		StatusCode: 404,
		Body:       io.NopCloser(strings.NewReader("")),
	}, nil
}

// TestRealAnalyzePage tests the realAnalyzePage function using mocked HTTP responses
func TestRealAnalyzePage(t *testing.T) {
	// Sample HTML string simulating a real webpage with:
	// - Doctype (<!DOCTYPE html>)
	// - Title tag with "Test Page"
	// - Headings h1 and h2
	// - Internal and external links
	// - A password input field (to detect login form)
	const testHTML = `<!DOCTYPE html>
<html>
<head><title>Test Page</title></head>
<body>
<h1>Main Heading</h1>
<h2>Sub Heading</h2>
<a href="/internal">Internal Link</a>
<a href="https://external.com/page">External Link</a>
<input type="password" name="pass"/>
</body>
</html>`

	// Base URL to be used as the test target URL for analysis
	baseURL := "https://example.com"

	// Create a mock HTTP client with custom Transport to intercept HTTP calls
	mockClient := &http.Client{
		Transport: &mockRoundTripper{
			// mockGet returns the testHTML for the baseURL, 404 otherwise
			mockGet: func(req *http.Request) *http.Response {
				if req.URL.String() == baseURL {
					return &http.Response{
						StatusCode: 200,
						Body:       io.NopCloser(strings.NewReader(testHTML)),
						Header:     make(http.Header),
					}
				}
				// Return 404 for any other GET request
				return &http.Response{
					StatusCode: 404,
					Body:       io.NopCloser(strings.NewReader("")),
					Header:     make(http.Header),
				}
			},
			// mockHead simulates accessibility checks of links:
			// - internal link is accessible (status 200)
			// - external link is inaccessible (status 404)
			mockHead: func(req *http.Request) *http.Response {
				if req.URL.String() == baseURL+"/internal" {
					return &http.Response{
						StatusCode: 200,
						Body:       io.NopCloser(strings.NewReader("")),
						Header:     make(http.Header),
					}
				} else if req.URL.String() == "https://external.com/page" {
					return &http.Response{
						StatusCode: 404,
						Body:       io.NopCloser(strings.NewReader("")),
						Header:     make(http.Header),
					}
				}
				// Default 404 for all other HEAD requests
				return &http.Response{
					StatusCode: 404,
					Body:       io.NopCloser(strings.NewReader("")),
					Header:     make(http.Header),
				}
			},
		},
	}

	// Override the package-level httpClient with our mock client for testing
	origClient := httpClient
	httpClient = mockClient
	// Restore the original httpClient after the test finishes
	defer func() { httpClient = origClient }()

	// Call the realAnalyzePage function using the mocked HTTP client and the test URL
	result, err := realAnalyzePage(baseURL)
	if err != nil {
		t.Fatalf("realAnalyzePage returned error: %v", err)
	}

	// Validate that the extracted page title matches the test HTML's title
	if result.Title != "Test Page" {
		t.Errorf("Title = %q; want %q", result.Title, "Test Page")
	}

	// Check that the detected HTML version is "HTML 5" based on the doctype
	if result.HTMLVersion != "HTML 5" {
		t.Errorf("HTMLVersion = %q; want %q", result.HTMLVersion, "HTML 5")
	}

	// Verify that the login form detection works (should detect password input)
	if !result.LoginForm {
		t.Error("LoginForm = false; want true")
	}

	// Confirm the counts of headings extracted from the HTML:
	// h1 count should be 1
	if got, want := result.Headings["H1"], 1; got != want {
		t.Errorf("Headings[H1] = %d; want %d", got, want)
	}
	// h2 count should be 1
	if got, want := result.Headings["H2"], 1; got != want {
		t.Errorf("Headings[H2] = %d; want %d", got, want)
	}

	// Check the number of internal links found (should be 1)
	if got, want := result.InternalLinks, 1; got != want {
		t.Errorf("InternalLinks = %d; want %d", got, want)
	}
	// Check the number of external links found (should be 1)
	if got, want := result.ExternalLinks, 1; got != want {
		t.Errorf("ExternalLinks = %d; want %d", got, want)
	}

	// Verify the count of inaccessible links (external link simulated as inaccessible)
	if got, want := result.InaccessibleLinks, 1; got != want {
		t.Errorf("InaccessibleLinks = %d; want %d", got, want)
	}
}
