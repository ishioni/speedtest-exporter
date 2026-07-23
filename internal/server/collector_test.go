package server

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/ishioni/speedtest-exporter/internal/speedtest"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

type fakeRunner struct {
	mu     sync.Mutex
	calls  int
	result speedtest.Result
	err    error
}

func (r *fakeRunner) Run(context.Context) (speedtest.Result, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls++
	return r.result, r.err
}

func (r *fakeRunner) callCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.calls
}

func TestCollectorCachesResult(t *testing.T) {
	t.Parallel()

	registry := prometheus.NewRegistry()
	packetLoss := 0.5
	runner := &fakeRunner{result: speedtest.Result{
		ServerID:                       1,
		DownloadBitsPerSec:             100,
		DownloadBytes:                  200,
		DownloadElapsedSeconds:         3,
		DownloadLatencyIQMMilliseconds: 4,
		UploadBytes:                    500,
		UploadElapsedSeconds:           6,
		UploadLatencyIQMMilliseconds:   7,
		PacketLossPercent:              &packetLoss,
		ISP:                            "Provider A",
	}}
	collector := NewCollector(runner, time.Second, time.Minute, registry)
	if err := collector.Update(context.Background()); err != nil {
		t.Fatalf("first Update() error = %v", err)
	}
	if err := collector.Update(context.Background()); err != nil {
		t.Fatalf("second Update() error = %v", err)
	}
	if runner.callCount() != 1 {
		t.Fatalf("Run() calls = %d, want 1", runner.callCount())
	}
	if got := testutil.ToFloat64(collector.up); got != 1 {
		t.Fatalf("speedtest_up = %v, want 1", got)
	}
	if got := testutil.ToFloat64(collector.downloadBytes); got != 200 {
		t.Fatalf("speedtest_download_bytes = %v, want 200", got)
	}
	if got := testutil.ToFloat64(collector.downloadLatencyIQM); got != 4 {
		t.Fatalf("speedtest_download_latency_iqm_milliseconds = %v, want 4", got)
	}
	if got := testutil.ToFloat64(collector.uploadLatencyIQM); got != 7 {
		t.Fatalf("speedtest_upload_latency_iqm_milliseconds = %v, want 7", got)
	}
	if got := testutil.ToFloat64(collector.packetLoss.WithLabelValues()); got != packetLoss {
		t.Fatalf("speedtest_packet_loss_percent = %v, want %v", got, packetLoss)
	}
	if got := testutil.ToFloat64(collector.ispInfo.WithLabelValues("Provider A")); got != 1 {
		t.Fatalf("speedtest_isp_info{isp=Provider A} = %v, want 1", got)
	}
}

func TestCollectorReplacesISPInfoSeries(t *testing.T) {
	t.Parallel()

	registry := prometheus.NewRegistry()
	collector := NewCollector(&fakeRunner{}, time.Second, time.Minute, registry)
	collector.setResult(speedtest.Result{ISP: "Provider A"})
	collector.setResult(speedtest.Result{ISP: "Provider B"})

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("Gather() error = %v", err)
	}
	for _, family := range families {
		if family.GetName() != "speedtest_isp_info" {
			continue
		}
		if len(family.Metric) != 1 || len(family.Metric[0].Label) != 1 || family.Metric[0].Label[0].GetValue() != "Provider B" {
			t.Fatalf("speedtest_isp_info = %#v, want only Provider B", family.Metric)
		}
		return
	}
	t.Fatal("speedtest_isp_info was not gathered")
}

func TestCollectorExposesFailure(t *testing.T) {
	t.Parallel()

	registry := prometheus.NewRegistry()
	collector := NewCollector(&fakeRunner{err: errors.New("unavailable")}, time.Second, 0, registry)
	collector.setResult(speedtest.Result{DownloadBitsPerSec: 100, ISP: "Provider A"})
	collector.lastRun.Set(42)
	if err := collector.Update(context.Background()); err == nil {
		t.Fatal("Update() error = nil, want runner error")
	}
	if got := testutil.ToFloat64(collector.up); got != 0 {
		t.Fatalf("speedtest_up = %v, want 0", got)
	}

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("Gather() error = %v", err)
	}
	for _, family := range families {
		switch family.GetName() {
		case "speedtest_download_bits_per_second", "speedtest_isp_info", "speedtest_last_run_timestamp_seconds":
			t.Fatalf("%s must be omitted for a failed run", family.GetName())
		}
	}
}
