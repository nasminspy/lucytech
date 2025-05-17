package main

import (
	"log/slog"         // Structured logger
	"lucytech/handler" // Custom package for request handlers
	"lucytech/metrics" // Custom package for Prometheus metrics
	"net/http"         // HTTP server
	"os"               // For accessing stdout

	"github.com/prometheus/client_golang/prometheus/promhttp" // Prometheus metrics
)

// initLogger initializes the structured logger using slog with info-level logging.
func initLogger() {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo, // Only logs info and above (info, warn, error)
	})
	slog.SetDefault(slog.New(handler)) // Set slog as the default logger
}

// main is the entry point of the application
func main() {
	initLogger() // Initialize logging
	slog.Info("Logger initialized")

	metrics.Init() // Register custom Prometheus metrics
	slog.Info("Metrics initialized")

	// Load the HTML template used by the handlers
	handler.LoadTemplates("templates/index.html")
	slog.Info("Templates loaded", "path", "templates/index.html")

	// Start Prometheus metrics server in a separate goroutine
	go func() {
		http.Handle("/metrics", promhttp.Handler()) // Metrics endpoint handler
		slog.Info("Starting metrics server", "addr", "localhost:6060/metrics")
		if err := http.ListenAndServe("localhost:6060", nil); err != nil {
			slog.Error("Metrics server failed", "error", err)
		}
	}()

	// Register the HTTP handlers for home and analyze routes
	http.HandleFunc("/", handler.HomeHandler)
	http.HandleFunc("/analyze", handler.AnalyzeHandler)

	// Start the main HTTP server
	slog.Info("Starting application", "addr", ":8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		slog.Error("App server failed", "error", err)
	}
}
