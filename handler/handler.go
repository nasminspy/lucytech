package handler

import (
	"html/template"
	"log/slog"
	"lucytech/metrics"
	"lucytech/parser"
	"net/http"
	"time"
)

type ResultData struct {
	HTMLVersion       string
	Title             string
	Headings          map[string]int
	InternalLinks     int
	ExternalLinks     int
	InaccessibleLinks int
	LoginForm         bool
}

type PageData struct {
	Result *ResultData
	Error  string
}

var tmpl = template.Must(template.ParseFiles("templates/index.html"))

// HomeHandler renders the input form
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer metrics.RequestDuration.WithLabelValues("/", r.Method).Observe(time.Since(start).Seconds())
	metrics.RequestCount.WithLabelValues("/", r.Method).Inc()

	if err := tmpl.Execute(w, PageData{}); err != nil {
		slog.Error("Failed to render home page", "error", err)
	}
}

// AnalyzeHandler processes the submitted URL
func AnalyzeHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer metrics.RequestDuration.WithLabelValues("/analyze", r.Method).Observe(time.Since(start).Seconds())
	metrics.RequestCount.WithLabelValues("/analyze", r.Method).Inc()

	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	url := r.FormValue("url")
	if url == "" {
		if err := tmpl.Execute(w, PageData{Error: "URL is required"}); err != nil {
			slog.Error("Failed to render error page", "error", err)
		}
		return
	}

	analysis, err := parser.AnalyzePage(url)
	if err != nil {
		slog.Error("Error analyzing page", "url", url, "error", err)
		if err := tmpl.Execute(w, PageData{Error: err.Error()}); err != nil {
			slog.Error("Failed to render error page", "error", err)
		}
		return
	}

	data := &ResultData{
		HTMLVersion:       analysis.HTMLVersion,
		Title:             analysis.Title,
		Headings:          analysis.Headings,
		InternalLinks:     analysis.InternalLinks,
		ExternalLinks:     analysis.ExternalLinks,
		InaccessibleLinks: analysis.InaccessibleLinks,
		LoginForm:         analysis.LoginForm,
	}

	if err := tmpl.Execute(w, PageData{Result: data}); err != nil {
		slog.Error("Failed to render analysis result", "error", err)
	}
}
