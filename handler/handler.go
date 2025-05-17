package handler

import (
	"html/template"
	"log/slog"
	"lucytech/metrics"
	"lucytech/parser"
	"net/http"
	"time"
)

// ResultData holds the analysis results that will be passed to the template for rendering.
type ResultData struct {
	HTMLVersion       string         // Detected HTML version of the analyzed page
	Title             string         // Page title
	Headings          map[string]int // Counts of headings by level (e.g., H1, H2)
	InternalLinks     int            // Number of internal links on the page
	ExternalLinks     int            // Number of external links on the page
	InaccessibleLinks int            // Number of links that were inaccessible (HTTP errors)
	LoginForm         bool           // Whether a login form with password field was detected
}

// PageData wraps ResultData or Error message to pass to the HTML template.
type PageData struct {
	Result *ResultData // Populated when analysis succeeds
	Error  string      // Populated when there's an error to display
}

var tmpl *template.Template

// LoadTemplates loads HTML templates from disk and sets the global tmpl variable.
// This should be called once during application startup to parse templates.
func LoadTemplates(path string) {
	// Parse template files from given path. Panic on error to fail fast on startup.
	tmpl = template.Must(template.ParseFiles(path))
	slog.Info("Templates loaded successfully", "path", path)
}

// HomeHandler serves the initial home page with the URL input form.
// Tracks request count and duration metrics for the "/" endpoint.
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	// Observe and record request duration for metrics
	defer metrics.RequestDuration.WithLabelValues("/", r.Method).Observe(time.Since(start).Seconds())
	// Increment request count metric for monitoring
	metrics.RequestCount.WithLabelValues("/", r.Method).Inc()

	slog.Debug("Serving home page")

	// Render template with empty PageData (no results or errors yet)
	if err := tmpl.Execute(w, PageData{}); err != nil {
		// Log template rendering errors with details
		slog.Error("Failed to render home page template", "error", err)
	}
}

// AnalyzeHandler processes the submitted URL from the form and returns analysis results.
// Tracks metrics, validates input, handles errors, and renders results or error messages.
func AnalyzeHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	// Track duration of /analyze requests for metrics
	defer metrics.RequestDuration.WithLabelValues("/analyze", r.Method).Observe(time.Since(start).Seconds())
	metrics.RequestCount.WithLabelValues("/analyze", r.Method).Inc()

	slog.Debug("AnalyzeHandler invoked", "method", r.Method)

	// Only allow POST method for analysis submission
	if r.Method != http.MethodPost {
		slog.Warn("Invalid HTTP method for analyze endpoint", "method", r.Method)
		http.Redirect(w, r, "/", http.StatusSeeOther) // Redirect GET or other methods back to home page
		return
	}

	// Extract submitted URL from form data
	url := r.FormValue("url")
	if url == "" {
		slog.Warn("No URL provided in form submission")
		// Render page with error message about missing URL
		if err := tmpl.Execute(w, PageData{Error: "URL is required"}); err != nil {
			slog.Error("Failed to render error message template", "error", err)
		}
		return
	}

	slog.Info("Starting page analysis", "url", url)

	// Call parser package to analyze the given URL
	analysis, err := parser.AnalyzePage(url)
	if err != nil {
		slog.Error("Page analysis failed", "url", url, "error", err)
		// Render page showing error to user
		if err := tmpl.Execute(w, PageData{Error: err.Error()}); err != nil {
			slog.Error("Failed to render error page after analysis failure", "error", err)
		}
		return
	}

	slog.Info("Page analysis successful", "url", url)

	// Prepare the results for rendering in template
	data := &ResultData{
		HTMLVersion:       analysis.HTMLVersion,
		Title:             analysis.Title,
		Headings:          analysis.Headings,
		InternalLinks:     analysis.InternalLinks,
		ExternalLinks:     analysis.ExternalLinks,
		InaccessibleLinks: analysis.InaccessibleLinks,
		LoginForm:         analysis.LoginForm,
	}

	// Render results page with analysis data
	if err := tmpl.Execute(w, PageData{Result: data}); err != nil {
		slog.Error("Failed to render analysis result template", "error", err)
	}
}
