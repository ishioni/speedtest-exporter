package server

import (
	"context"
	"sync"
	"time"

	"github.com/ishioni/speedtest-exporter/internal/speedtest"
	"github.com/prometheus/client_golang/prometheus"
)

// Collector updates the Speedtest metrics, serializing expensive CLI runs and
// optionally reusing the most recent outcome for a configured TTL.
type Collector struct {
	runner   speedtest.Runner
	timeout  time.Duration
	cacheFor time.Duration
	now      func() time.Time

	mu           sync.Mutex
	hasResult    bool
	cacheExpires time.Time
	currentISP   string

	serverID              prometheus.Gauge
	jitter                prometheus.Gauge
	ping                  prometheus.Gauge
	pingLow               prometheus.Gauge
	pingHigh              prometheus.Gauge
	download              prometheus.Gauge
	downloadBytes         prometheus.Gauge
	downloadElapsed       prometheus.Gauge
	downloadLatencyIQM    prometheus.Gauge
	downloadLatencyLow    prometheus.Gauge
	downloadLatencyHigh   prometheus.Gauge
	downloadLatencyJitter prometheus.Gauge
	upload                prometheus.Gauge
	uploadBytes           prometheus.Gauge
	uploadElapsed         prometheus.Gauge
	uploadLatencyIQM      prometheus.Gauge
	uploadLatencyLow      prometheus.Gauge
	uploadLatencyHigh     prometheus.Gauge
	uploadLatencyJitter   prometheus.Gauge
	packetLoss            *prometheus.GaugeVec
	ispInfo               *prometheus.GaugeVec
	up                    prometheus.Gauge
	runTotal              *prometheus.CounterVec
	duration              prometheus.Histogram
	lastRun               prometheus.Gauge
}

// NewCollector registers the exporter's metrics with registerer.
func NewCollector(
	runner speedtest.Runner,
	timeout, cacheFor time.Duration,
	registerer prometheus.Registerer,
) *Collector {
	collector := &Collector{
		runner:   runner,
		timeout:  timeout,
		cacheFor: cacheFor,
		now:      time.Now,
		serverID: gauge("speedtest_server_id", "Speedtest server ID used by the most recent test."),
		jitter: gauge(
			"speedtest_jitter_latency_milliseconds",
			"Jitter from the most recent Speedtest measurement in milliseconds.",
		),
		ping: gauge(
			"speedtest_ping_latency_milliseconds",
			"Ping latency from the most recent Speedtest measurement in milliseconds.",
		),
		pingLow: gauge(
			"speedtest_ping_low_latency_milliseconds",
			"Lowest ping latency from the most recent Speedtest measurement in milliseconds.",
		),
		pingHigh: gauge(
			"speedtest_ping_high_latency_milliseconds",
			"Highest ping latency from the most recent Speedtest measurement in milliseconds.",
		),
		download: gauge(
			"speedtest_download_bits_per_second",
			"Download bandwidth from the most recent Speedtest measurement in bits per second.",
		),
		downloadBytes: gauge(
			"speedtest_download_bytes",
			"Bytes transferred during the most recent Speedtest download phase.",
		),
		downloadElapsed: gauge(
			"speedtest_download_elapsed_seconds",
			"Elapsed time of the most recent Speedtest download phase in seconds.",
		),
		downloadLatencyIQM: gauge(
			"speedtest_download_latency_iqm_milliseconds",
			"Interquartile mean latency during the most recent Speedtest download phase in milliseconds.",
		),
		downloadLatencyLow: gauge(
			"speedtest_download_latency_low_milliseconds",
			"Lowest latency during the most recent Speedtest download phase in milliseconds.",
		),
		downloadLatencyHigh: gauge(
			"speedtest_download_latency_high_milliseconds",
			"Highest latency during the most recent Speedtest download phase in milliseconds.",
		),
		downloadLatencyJitter: gauge(
			"speedtest_download_latency_jitter_milliseconds",
			"Jitter during the most recent Speedtest download phase in milliseconds.",
		),
		upload: gauge(
			"speedtest_upload_bits_per_second",
			"Upload bandwidth from the most recent Speedtest measurement in bits per second.",
		),
		uploadBytes: gauge(
			"speedtest_upload_bytes",
			"Bytes transferred during the most recent Speedtest upload phase.",
		),
		uploadElapsed: gauge(
			"speedtest_upload_elapsed_seconds",
			"Elapsed time of the most recent Speedtest upload phase in seconds.",
		),
		uploadLatencyIQM: gauge(
			"speedtest_upload_latency_iqm_milliseconds",
			"Interquartile mean latency during the most recent Speedtest upload phase in milliseconds.",
		),
		uploadLatencyLow: gauge(
			"speedtest_upload_latency_low_milliseconds",
			"Lowest latency during the most recent Speedtest upload phase in milliseconds.",
		),
		uploadLatencyHigh: gauge(
			"speedtest_upload_latency_high_milliseconds",
			"Highest latency during the most recent Speedtest upload phase in milliseconds.",
		),
		uploadLatencyJitter: gauge(
			"speedtest_upload_latency_jitter_milliseconds",
			"Jitter during the most recent Speedtest upload phase in milliseconds.",
		),
		packetLoss: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "speedtest_packet_loss_percent",
			Help: "Packet loss reported by the most recent Speedtest measurement in percent.",
		}, nil),
		ispInfo: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "speedtest_isp_info",
			Help: "Internet service provider reported by the most recent successful Speedtest measurement.",
		}, []string{"isp"}),
		up: gauge(
			"speedtest_up",
			"Whether the most recent Speedtest measurement succeeded (1) or failed (0).",
		),
		runTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "speedtest_runs_total",
			Help: "Total Speedtest CLI runs, partitioned by outcome.",
		}, []string{"result"}),
		duration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "speedtest_run_duration_seconds",
			Help:    "Wall-clock duration of a Speedtest CLI run.",
			Buckets: []float64{5, 15, 30, 60, 90, 120, 180},
		}),
		lastRun: gauge(
			"speedtest_last_run_timestamp_seconds",
			"Unix timestamp when the most recent Speedtest CLI run completed.",
		),
	}
	registerer.MustRegister(
		collector.serverID,
		collector.jitter,
		collector.ping,
		collector.pingLow,
		collector.pingHigh,
		collector.download,
		collector.downloadBytes,
		collector.downloadElapsed,
		collector.downloadLatencyIQM,
		collector.downloadLatencyLow,
		collector.downloadLatencyHigh,
		collector.downloadLatencyJitter,
		collector.upload,
		collector.uploadBytes,
		collector.uploadElapsed,
		collector.uploadLatencyIQM,
		collector.uploadLatencyLow,
		collector.uploadLatencyHigh,
		collector.uploadLatencyJitter,
		collector.packetLoss,
		collector.ispInfo,
		collector.up,
		collector.runTotal,
		collector.duration,
		collector.lastRun,
	)
	return collector
}

func gauge(name, help string) prometheus.Gauge {
	return prometheus.NewGauge(prometheus.GaugeOpts{Name: name, Help: help})
}

// Update ensures metrics contain a fresh measurement. Failures are reflected in
// speedtest_up and still leave the scrape itself successful, so Prometheus can
// alert on the measurement rather than an inaccessible target.
func (c *Collector) Update(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := c.now()
	if c.hasResult && now.Before(c.cacheExpires) {
		return nil
	}

	runCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	started := c.now()
	result, err := c.runner.Run(runCtx)
	completed := c.now()
	c.duration.Observe(completed.Sub(started).Seconds())
	c.lastRun.Set(float64(completed.UnixNano()) / float64(time.Second))
	c.hasResult = true
	c.cacheExpires = completed.Add(c.cacheFor)

	if err != nil {
		c.setFailure()
		c.runTotal.WithLabelValues("error").Inc()
		return err
	}
	c.setResult(result)
	c.runTotal.WithLabelValues("success").Inc()
	return nil
}

func (c *Collector) setResult(result speedtest.Result) {
	c.setISP(result.ISP)
	c.serverID.Set(result.ServerID)
	c.jitter.Set(result.JitterMilliseconds)
	c.ping.Set(result.PingMilliseconds)
	c.pingLow.Set(result.PingLowMilliseconds)
	c.pingHigh.Set(result.PingHighMilliseconds)
	c.download.Set(result.DownloadBitsPerSec)
	c.downloadBytes.Set(result.DownloadBytes)
	c.downloadElapsed.Set(result.DownloadElapsedSeconds)
	c.downloadLatencyIQM.Set(result.DownloadLatencyIQMMilliseconds)
	c.downloadLatencyLow.Set(result.DownloadLatencyLowMilliseconds)
	c.downloadLatencyHigh.Set(result.DownloadLatencyHighMilliseconds)
	c.downloadLatencyJitter.Set(result.DownloadLatencyJitterMilliseconds)
	c.upload.Set(result.UploadBitsPerSecond)
	c.uploadBytes.Set(result.UploadBytes)
	c.uploadElapsed.Set(result.UploadElapsedSeconds)
	c.uploadLatencyIQM.Set(result.UploadLatencyIQMMilliseconds)
	c.uploadLatencyLow.Set(result.UploadLatencyLowMilliseconds)
	c.uploadLatencyHigh.Set(result.UploadLatencyHighMilliseconds)
	c.uploadLatencyJitter.Set(result.UploadLatencyJitterMilliseconds)
	if result.PacketLossPercent == nil {
		c.packetLoss.DeleteLabelValues()
	} else {
		c.packetLoss.WithLabelValues().Set(*result.PacketLossPercent)
	}
	c.up.Set(1)
}

func (c *Collector) setISP(isp string) {
	if isp == "" {
		return
	}
	if c.currentISP != "" && c.currentISP != isp {
		c.ispInfo.DeleteLabelValues(c.currentISP)
	}
	c.ispInfo.WithLabelValues(isp).Set(1)
	c.currentISP = isp
}

func (c *Collector) setFailure() {
	c.serverID.Set(0)
	c.jitter.Set(0)
	c.ping.Set(0)
	c.pingLow.Set(0)
	c.pingHigh.Set(0)
	c.download.Set(0)
	c.downloadBytes.Set(0)
	c.downloadElapsed.Set(0)
	c.downloadLatencyIQM.Set(0)
	c.downloadLatencyLow.Set(0)
	c.downloadLatencyHigh.Set(0)
	c.downloadLatencyJitter.Set(0)
	c.upload.Set(0)
	c.uploadBytes.Set(0)
	c.uploadElapsed.Set(0)
	c.uploadLatencyIQM.Set(0)
	c.uploadLatencyLow.Set(0)
	c.uploadLatencyHigh.Set(0)
	c.uploadLatencyJitter.Set(0)
	c.packetLoss.DeleteLabelValues()
	c.up.Set(0)
}
