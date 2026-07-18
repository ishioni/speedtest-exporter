// Package server exposes the exporter HTTP endpoints and Prometheus registry.
package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/ishioni/speedtest-exporter/internal/config"
	"github.com/ishioni/speedtest-exporter/internal/speedtest"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Server owns the exporter HTTP listener and its metrics registry.
type Server struct {
	config    config.Config
	collector *Collector
	logger    *slog.Logger
	handler   http.Handler
}

// New creates a ready-to-run Speedtest exporter server.
func New(cfg config.Config, runner speedtest.Runner, logger *slog.Logger) *Server {
	registry := prometheus.NewRegistry()
	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)
	collector := NewCollector(runner, cfg.Timeout, cfg.CacheFor, registry)

	server := &Server{
		config:    cfg,
		collector: collector,
		logger:    logger,
	}
	server.handler = server.routes(promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	return server
}

// Run serves until ctx is cancelled or the listener returns an unexpected error.
func (s *Server) Run(ctx context.Context) error {
	listener, err := net.Listen("tcp", s.config.ListenAddress)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", s.config.ListenAddress, err)
	}
	defer func() { _ = listener.Close() }()

	httpServer := &http.Server{
		Handler:           s.handler,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	errCh := make(chan error, 1)
	go func() {
		errCh <- httpServer.Serve(listener)
	}()

	s.logger.Info("HTTP server listening", "address", s.config.ListenAddress)
	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("serve HTTP: %w", err)
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown HTTP server: %w", err)
		}
		return nil
	}
}

func (s *Server) routes(metricsHandler http.Handler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = writer.Write([]byte("speedtest-exporter\n"))
	})
	mux.HandleFunc("GET /healthz", func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("GET /readyz", func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusOK)
	})
	mux.Handle("GET /metrics", http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if err := s.collector.Update(request.Context()); err != nil {
			s.logger.Error("Speedtest CLI run failed", "error", err)
		}
		metricsHandler.ServeHTTP(writer, request)
	}))
	return mux
}
