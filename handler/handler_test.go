package handler

import (
	"errors"
	"html/template"
	"lucytech/parser"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// Initialize the HTML template used for rendering responses in tests
func init() {
	// Define a very simple template that either shows an error message or the page title from analysis
	tmpl = template.Must(template.New("index").Parse(`
		{{if .Error}}Error: {{.Error}}{{else}}Title: {{.Result.Title}}{{end}}
	`))
}

// TestHomeHandler checks that the HomeHandler returns a successful HTTP 200 status
func TestHomeHandler(t *testing.T) {
	// Create a new GET request to the root URL "/"
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	// Create a ResponseRecorder to capture the response
	w := httptest.NewRecorder()

	// Call the HomeHandler with the request and recorder
	HomeHandler(w, req)

	// Get the response from the recorder
	resp := w.Result()
	// Verify that the status code is 200 OK
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

// TestAnalyzeHandler_GetRedirects verifies that GET requests to /analyze redirect (HTTP 303)
func TestAnalyzeHandler_GetRedirects(t *testing.T) {
	// Create a GET request to /analyze
	req := httptest.NewRequest(http.MethodGet, "/analyze", nil)
	w := httptest.NewRecorder()

	// Call AnalyzeHandler with GET method
	AnalyzeHandler(w, req)

	// Expect HTTP status 303 See Other for GET requests (redirect to home or form)
	if w.Result().StatusCode != http.StatusSeeOther {
		t.Errorf("expected redirect, got %d", w.Result().StatusCode)
	}
}

// TestAnalyzeHandler_EmptyURL checks that posting without a URL returns a relevant error
func TestAnalyzeHandler_EmptyURL(t *testing.T) {
	// Prepare an empty form (no URL parameter)
	form := url.Values{}
	req := httptest.NewRequest(http.MethodPost, "/analyze", strings.NewReader(form.Encode()))
	// Set content type for form submission
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	// Call AnalyzeHandler with POST and empty form data
	AnalyzeHandler(w, req)

	// Expect an error message mentioning "URL is required"
	if !strings.Contains(w.Body.String(), "URL is required") {
		t.Errorf("expected error for missing URL, got %s", w.Body.String())
	}
}

// TestAnalyzeHandler_ValidURL tests the handler behavior on a valid URL input with mocked parser
func TestAnalyzeHandler_ValidURL(t *testing.T) {
	// Mock the AnalyzePage function in parser package to return a fixed result without making HTTP calls
	parser.AnalyzePage = func(url string) (*parser.AnalysisResult, error) {
		return &parser.AnalysisResult{
			HTMLVersion:       "HTML5",
			Title:             "Test Title",
			Headings:          map[string]int{"h1": 1},
			InternalLinks:     2,
			ExternalLinks:     3,
			InaccessibleLinks: 0,
			LoginForm:         false,
		}, nil
	}

	// Create a form with a valid URL parameter
	form := url.Values{}
	form.Set("url", "http://example.com")
	req := httptest.NewRequest(http.MethodPost, "/analyze", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	// Call AnalyzeHandler with POST and valid URL
	AnalyzeHandler(w, req)

	// Verify that the response body contains the mocked title from the analysis
	if !strings.Contains(w.Body.String(), "Test Title") {
		t.Errorf("expected analysis result title, got %s", w.Body.String())
	}
}

// TestAnalyzeHandler_ErrorFromParser verifies the handler handles parser errors gracefully
func TestAnalyzeHandler_ErrorFromParser(t *testing.T) {
	// Mock AnalyzePage to return an error simulating a failure in parsing the URL
	parser.AnalyzePage = func(url string) (*parser.AnalysisResult, error) {
		return nil, errors.New("mock parse error")
	}

	// Create form data with a valid URL (though parser is mocked to fail)
	form := url.Values{}
	form.Set("url", "http://invalid.com")
	req := httptest.NewRequest(http.MethodPost, "/analyze", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	// Call AnalyzeHandler
	AnalyzeHandler(w, req)

	// Expect the response to contain the mock error message
	if !strings.Contains(w.Body.String(), "mock parse error") {
		t.Errorf("expected mock parse error, got %s", w.Body.String())
	}
}
