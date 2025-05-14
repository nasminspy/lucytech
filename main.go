package main

import (
	"log/slog"
	"net/http"
	"os"

	"lucytech/handler"
	"lucytech/metrics"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func initLogger() {
	// You can switch to slog.NewJSONHandler if you prefer plain logs over JSON
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	slog.SetDefault(slog.New(handler))
}

func main() {
	initLogger()

	metrics.Init()

	// Start Prometheus metrics server in a separate goroutine
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		slog.Info("Starting metrics server", "addr", "localhost:6060/metrics")
		if err := http.ListenAndServe("localhost:6060", nil); err != nil {
			slog.Error("Metrics server failed", "error", err)
		}
	}()

	// Register main handlers
	http.HandleFunc("/", handler.HomeHandler)
	http.HandleFunc("/analyze", handler.AnalyzeHandler)

	slog.Info("Starting application", "addr", ":8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		slog.Error("App server failed", "error", err)
	}
}
