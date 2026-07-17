package speedtest

import (
	"strings"
	"testing"
)

const completeResult = `{
	"type": "result",
	"packetLoss": 0.5,
	"isp": "Example ISP",
	"server": {"id": 12345},
	"ping": {"jitter": 1.25, "latency": 12.5, "low": 10, "high": 15},
	"download": {
		"bandwidth": 12500000,
		"bytes": 150000000,
		"elapsed": 12000,
		"latency": {"iqm": 20, "low": 12, "high": 45, "jitter": 5}
	},
	"upload": {
		"bandwidth": 2500000,
		"bytes": 30000000,
		"elapsed": 15000,
		"latency": {"iqm": 30, "low": 15, "high": 60, "jitter": 8}
	}
}`

func TestParse(t *testing.T) {
	t.Parallel()

	result, err := Parse([]byte(completeResult))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if result.ServerID != 12345 || result.ISP != "Example ISP" || result.JitterMilliseconds != 1.25 || result.PingMilliseconds != 12.5 {
		t.Fatalf("Parse() basic result = %#v", result)
	}
	if result.PingLowMilliseconds != 10 || result.PingHighMilliseconds != 15 {
		t.Fatalf("Parse() ping range = %#v", result)
	}
	if result.DownloadBitsPerSec != 100000000 || result.UploadBitsPerSecond != 20000000 {
		t.Fatalf("Parse() bandwidth conversion = %#v", result)
	}
	if result.DownloadBytes != 150000000 || result.DownloadElapsedSeconds != 12 {
		t.Fatalf("Parse() download phase = %#v", result)
	}
	if result.UploadBytes != 30000000 || result.UploadElapsedSeconds != 15 {
		t.Fatalf("Parse() upload phase = %#v", result)
	}
	if result.DownloadLatencyIQMMilliseconds != 20 || result.UploadLatencyJitterMilliseconds != 8 {
		t.Fatalf("Parse() loaded latency = %#v", result)
	}
	if result.PacketLossPercent == nil || *result.PacketLossPercent != 0.5 {
		t.Fatalf("Parse() packet loss = %#v", result.PacketLossPercent)
	}
}

func TestParseAllowsMissingPacketLoss(t *testing.T) {
	t.Parallel()

	result, err := Parse([]byte(strings.Replace(completeResult, `"packetLoss": 0.5,`, "", 1)))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if result.PacketLossPercent != nil {
		t.Fatalf("PacketLossPercent = %v, want nil", *result.PacketLossPercent)
	}
}

func TestParseAcceptsLeadingCLINotices(t *testing.T) {
	t.Parallel()

	result, err := Parse([]byte("License acceptance recorded. Continuing.\n" + completeResult))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if result.ServerID != 12345 {
		t.Fatalf("Parse() result = %#v", result)
	}
}

func TestParseRejectsCLIError(t *testing.T) {
	t.Parallel()

	_, err := Parse([]byte(`{"error": "network unavailable"}`))
	if err == nil || !strings.Contains(err.Error(), "network unavailable") {
		t.Fatalf("Parse() error = %v, want CLI error", err)
	}
}

func TestParseRejectsMissingFields(t *testing.T) {
	t.Parallel()

	_, err := Parse([]byte(`{"type": "result"}`))
	if err == nil || !strings.Contains(err.Error(), "server.id") {
		t.Fatalf("Parse() error = %v, want server.id error", err)
	}
}
