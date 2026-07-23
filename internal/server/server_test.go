package server

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ishioni/speedtest-exporter/internal/config"
	"github.com/ishioni/speedtest-exporter/internal/speedtest"
)

func TestMetricsRunsCollectorAndServesResult(t *testing.T) {
	t.Parallel()

	runner := &fakeRunner{result: speedtest.Result{ServerID: 42, DownloadBitsPerSec: 100}}
	srv := New(config.Config{Timeout: time.Second}, runner, slog.New(slog.NewTextHandler(io.Discard, nil)))
	response := httptest.NewRecorder()
	srv.handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("GET /metrics status = %d, want 200", response.Code)
	}
	body := response.Body.String()
	if !strings.Contains(body, "speedtest_server_id 42") || !strings.Contains(body, "speedtest_up 1") {
		t.Fatalf("GET /metrics body missing Speedtest metrics:\n%s", body)
	}
	if runner.callCount() != 1 {
		t.Fatalf("Run() calls = %d, want 1", runner.callCount())
	}
}

func TestMetricsLogsRawCLIOutputAtDebug(t *testing.T) {
	t.Parallel()

	binary := filepath.Join(t.TempDir(), "speedtest")
	const cliOutput = `{"type":"log","message":"temporary Ookla failure"}`
	if err := os.WriteFile(binary, []byte("#!/bin/sh\nprintf '%s\\n' '"+cliOutput+"'\n"), 0o755); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	var logs bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logs, &slog.HandlerOptions{Level: slog.LevelDebug}))
	srv := New(config.Config{Timeout: time.Second}, speedtest.NewClient(binary, ""), logger)
	response := httptest.NewRecorder()
	srv.handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("GET /metrics status = %d, want 200", response.Code)
	}
	if !strings.Contains(logs.String(), "Speedtest CLI output") || !strings.Contains(logs.String(), "temporary Ookla failure") {
		t.Fatalf("debug logs did not contain raw CLI output:\n%s", logs.String())
	}
}

func TestHealthEndpoints(t *testing.T) {
	t.Parallel()

	srv := New(config.Config{Timeout: time.Second}, &fakeRunner{}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	for _, path := range []string{"/", "/healthz", "/readyz"} {
		response := httptest.NewRecorder()
		srv.handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusOK {
			t.Errorf("GET %s status = %d, want 200", path, response.Code)
		}
	}
}
