// Command speedtest-exporter exposes official Ookla Speedtest measurements to Prometheus.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/ishioni/speedtest-exporter/internal/config"
	"github.com/ishioni/speedtest-exporter/internal/server"
	"github.com/ishioni/speedtest-exporter/internal/speedtest"
)

// Build metadata is injected with -ldflags during image and release builds.
var (
	version = "dev"
	commit  = "unknown"
)

func main() {
	if err := run(); err != nil {
		slog.Error("speedtest-exporter exited", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	logger := newLogger(cfg.LogLevel, cfg.LogFormat)
	slog.SetDefault(logger)
	client := speedtest.NewClient(cfg.Binary, cfg.ServerID)

	verifyCtx, cancelVerify := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelVerify()
	if err := client.Verify(verifyCtx); err != nil {
		return fmt.Errorf("verify Speedtest CLI: %w", err)
	}

	logger.Info("starting speedtest-exporter",
		"version", version,
		"commit", commit,
		"address", cfg.ListenAddress,
		"timeout", cfg.Timeout,
		"cache_for", cfg.CacheFor,
		"server_id", cfg.ServerID,
	)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		stop() // Restore default signal handling while graceful shutdown is in progress.
	}()

	return server.New(cfg, client, logger).Run(ctx)
}

func newLogger(level, format string) *slog.Logger {
	options := &slog.HandlerOptions{Level: parseLevel(level)}
	if strings.EqualFold(format, "text") {
		return slog.New(slog.NewTextHandler(os.Stdout, options))
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, options))
}

func parseLevel(value string) slog.Level {
	switch strings.ToLower(value) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
