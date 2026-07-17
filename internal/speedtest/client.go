// Package speedtest executes and parses the official Ookla Speedtest CLI.
package speedtest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// Result is a successful Speedtest measurement in Prometheus base units.
type Result struct {
	ServerID                          float64
	ISP                               string
	JitterMilliseconds                float64
	PingMilliseconds                  float64
	PingLowMilliseconds               float64
	PingHighMilliseconds              float64
	DownloadBitsPerSec                float64
	DownloadBytes                     float64
	DownloadElapsedSeconds            float64
	DownloadLatencyIQMMilliseconds    float64
	DownloadLatencyLowMilliseconds    float64
	DownloadLatencyHighMilliseconds   float64
	DownloadLatencyJitterMilliseconds float64
	UploadBitsPerSecond               float64
	UploadBytes                       float64
	UploadElapsedSeconds              float64
	UploadLatencyIQMMilliseconds      float64
	UploadLatencyLowMilliseconds      float64
	UploadLatencyHighMilliseconds     float64
	UploadLatencyJitterMilliseconds   float64
	// PacketLossPercent is nil when the CLI does not provide packet-loss data.
	PacketLossPercent *float64
}

// Runner produces one Speedtest result.
type Runner interface {
	Run(context.Context) (Result, error)
}

// Client runs the official Ookla Speedtest CLI.
type Client struct {
	binary   string
	serverID string
}

// NewClient returns a client that invokes binary. serverID is optional.
func NewClient(binary, serverID string) *Client {
	return &Client{binary: binary, serverID: serverID}
}

// Verify confirms the configured executable is the official Ookla CLI.
func (c *Client) Verify(ctx context.Context) error {
	output, err := exec.CommandContext(ctx, c.binary, "--version").CombinedOutput()
	if err != nil {
		return fmt.Errorf("run %q --version: %w", c.binary, err)
	}
	if !bytes.Contains(output, []byte("Speedtest by Ookla")) {
		return fmt.Errorf("%q is not the official Ookla Speedtest CLI", c.binary)
	}
	return nil
}

// Run starts a single Speedtest measurement and parses its JSON result.
func (c *Client) Run(ctx context.Context) (Result, error) {
	args := []string{"--format=json", "--progress=no", "--accept-license", "--accept-gdpr"}
	if c.serverID != "" {
		args = append(args, "--server-id="+c.serverID)
	}

	output, err := exec.CommandContext(ctx, c.binary, args...).CombinedOutput()
	if err != nil {
		return Result{}, fmt.Errorf("run speedtest: %w: %s", err, truncate(strings.TrimSpace(string(output)), 512))
	}
	return Parse(output)
}

// Parse converts the JSON emitted by the official CLI to Prometheus base units.
func Parse(output []byte) (Result, error) {
	// On a fresh HOME directory, the official CLI writes license/GDPR acceptance
	// notices before the JSON result despite --accept-license/--accept-gdpr.
	// Decode from the first JSON object instead of assuming stdout is JSON-only.
	jsonStart := bytes.IndexByte(output, '{')
	if jsonStart < 0 {
		return Result{}, fmt.Errorf("speedtest output did not contain a JSON result")
	}

	var raw cliResult
	decoder := json.NewDecoder(bytes.NewReader(output[jsonStart:]))
	decoder.UseNumber()
	if err := decoder.Decode(&raw); err != nil {
		return Result{}, fmt.Errorf("decode speedtest JSON: %w", err)
	}
	if len(raw.Error) > 0 && string(raw.Error) != "null" && string(raw.Error) != `""` {
		return Result{}, fmt.Errorf("speedtest reported an error: %s", truncate(string(raw.Error), 256))
	}
	if raw.Type != "" && raw.Type != "result" {
		return Result{}, fmt.Errorf("unexpected speedtest JSON type %q", raw.Type)
	}

	serverID, err := parseNumber(raw.Server.ID)
	if err != nil {
		return Result{}, fmt.Errorf("server.id: %w", err)
	}
	jitter, err := parseNumber(raw.Ping.Jitter)
	if err != nil {
		return Result{}, fmt.Errorf("ping.jitter: %w", err)
	}
	latency, err := parseNumber(raw.Ping.Latency)
	if err != nil {
		return Result{}, fmt.Errorf("ping.latency: %w", err)
	}
	pingLow, err := parseNumber(raw.Ping.Low)
	if err != nil {
		return Result{}, fmt.Errorf("ping.low: %w", err)
	}
	pingHigh, err := parseNumber(raw.Ping.High)
	if err != nil {
		return Result{}, fmt.Errorf("ping.high: %w", err)
	}
	download, err := parsePhase(raw.Download, "download")
	if err != nil {
		return Result{}, err
	}
	upload, err := parsePhase(raw.Upload, "upload")
	if err != nil {
		return Result{}, err
	}
	packetLoss, err := parseOptionalNumber(raw.PacketLoss)
	if err != nil {
		return Result{}, fmt.Errorf("packetLoss: %w", err)
	}

	return Result{
		ServerID:                          serverID,
		ISP:                               strings.TrimSpace(raw.ISP),
		JitterMilliseconds:                jitter,
		PingMilliseconds:                  latency,
		PingLowMilliseconds:               pingLow,
		PingHighMilliseconds:              pingHigh,
		DownloadBitsPerSec:                download.bandwidth * 8,
		DownloadBytes:                     download.bytes,
		DownloadElapsedSeconds:            download.elapsed / 1000,
		DownloadLatencyIQMMilliseconds:    download.latencyIQM,
		DownloadLatencyLowMilliseconds:    download.latencyLow,
		DownloadLatencyHighMilliseconds:   download.latencyHigh,
		DownloadLatencyJitterMilliseconds: download.latencyJitter,
		UploadBitsPerSecond:               upload.bandwidth * 8,
		UploadBytes:                       upload.bytes,
		UploadElapsedSeconds:              upload.elapsed / 1000,
		UploadLatencyIQMMilliseconds:      upload.latencyIQM,
		UploadLatencyLowMilliseconds:      upload.latencyLow,
		UploadLatencyHighMilliseconds:     upload.latencyHigh,
		UploadLatencyJitterMilliseconds:   upload.latencyJitter,
		PacketLossPercent:                 packetLoss,
	}, nil
}

type cliResult struct {
	Type       string          `json:"type"`
	Error      json.RawMessage `json:"error"`
	PacketLoss json.RawMessage `json:"packetLoss"`
	ISP        string          `json:"isp"`
	Server     struct {
		ID json.RawMessage `json:"id"`
	} `json:"server"`
	Ping struct {
		Jitter  json.RawMessage `json:"jitter"`
		Latency json.RawMessage `json:"latency"`
		Low     json.RawMessage `json:"low"`
		High    json.RawMessage `json:"high"`
	} `json:"ping"`
	Download cliPhase `json:"download"`
	Upload   cliPhase `json:"upload"`
}

type cliPhase struct {
	Bandwidth json.RawMessage `json:"bandwidth"`
	Bytes     json.RawMessage `json:"bytes"`
	Elapsed   json.RawMessage `json:"elapsed"`
	Latency   struct {
		IQM    json.RawMessage `json:"iqm"`
		Low    json.RawMessage `json:"low"`
		High   json.RawMessage `json:"high"`
		Jitter json.RawMessage `json:"jitter"`
	} `json:"latency"`
}

type phaseResult struct {
	bandwidth     float64
	bytes         float64
	elapsed       float64
	latencyIQM    float64
	latencyLow    float64
	latencyHigh   float64
	latencyJitter float64
}

func parsePhase(raw cliPhase, name string) (phaseResult, error) {
	bandwidth, err := parseNumber(raw.Bandwidth)
	if err != nil {
		return phaseResult{}, fmt.Errorf("%s.bandwidth: %w", name, err)
	}
	bytes, err := parseNumber(raw.Bytes)
	if err != nil {
		return phaseResult{}, fmt.Errorf("%s.bytes: %w", name, err)
	}
	elapsed, err := parseNumber(raw.Elapsed)
	if err != nil {
		return phaseResult{}, fmt.Errorf("%s.elapsed: %w", name, err)
	}
	iqm, err := parseNumber(raw.Latency.IQM)
	if err != nil {
		return phaseResult{}, fmt.Errorf("%s.latency.iqm: %w", name, err)
	}
	low, err := parseNumber(raw.Latency.Low)
	if err != nil {
		return phaseResult{}, fmt.Errorf("%s.latency.low: %w", name, err)
	}
	high, err := parseNumber(raw.Latency.High)
	if err != nil {
		return phaseResult{}, fmt.Errorf("%s.latency.high: %w", name, err)
	}
	jitter, err := parseNumber(raw.Latency.Jitter)
	if err != nil {
		return phaseResult{}, fmt.Errorf("%s.latency.jitter: %w", name, err)
	}
	return phaseResult{bandwidth, bytes, elapsed, iqm, low, high, jitter}, nil
}

func parseNumber(raw json.RawMessage) (float64, error) {
	value := strings.Trim(string(raw), `"`)
	if value == "" || value == "null" {
		return 0, fmt.Errorf("missing value")
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("parse %q: %w", value, err)
	}
	return parsed, nil
}

func parseOptionalNumber(raw json.RawMessage) (*float64, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	value, err := parseNumber(raw)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func truncate(value string, length int) string {
	if len(value) <= length {
		return value
	}
	return value[:length] + "…"
}
