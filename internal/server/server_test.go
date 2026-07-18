package server

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
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
