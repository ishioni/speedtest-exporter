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

	serverID              removableGauge
	jitter                removableGauge
	ping                  removableGauge
	pingLow               removableGauge
	pingHigh              removableGauge
	download              removableGauge
	downloadBytes         removableGauge
	downloadElapsed       removableGauge
	downloadLatencyIQM    removableGauge
	downloadLatencyLow    removableGauge
	downloadLatencyHigh   removableGauge
	downloadLatencyJitter removableGauge
	upload                removableGauge
	uploadBytes           removableGauge
	uploadElapsed         removableGauge
	uploadLatencyIQM      removableGauge
	uploadLatencyLow      removableGauge
	uploadLatencyHigh     removableGauge
	uploadLatencyJitter   removableGauge
	packetLoss            *prometheus.GaugeVec
	ispInfo               *prometheus.GaugeVec
	up                    prometheus.Gauge
	runTotal              *prometheus.CounterVec
	duration              prometheus.Histogram
	lastRun               removableGauge
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
		serverID: newRemovableGauge("speedtest_server_id", "Speedtest server ID used by the most recent successful test."),
		jitter: newRemovableGauge(
			"speedtest_jitter_latency_milliseconds",
			"Jitter from the most recent Speedtest measurement in milliseconds.",
		),
		ping: newRemovableGauge(
			"speedtest_ping_latency_milliseconds",
			"Ping latency from the most recent Speedtest measurement in milliseconds.",
		),
		pingLow: newRemovableGauge(
			"speedtest_ping_low_latency_milliseconds",
			"Lowest ping latency from the most recent Speedtest measurement in milliseconds.",
		),
		pingHigh: newRemovableGauge(
			"speedtest_ping_high_latency_milliseconds",
			"Highest ping latency from the most recent Speedtest measurement in milliseconds.",
		),
		download: newRemovableGauge(
			"speedtest_download_bits_per_second",
			"Download bandwidth from the most recent Speedtest measurement in bits per second.",
		),
		downloadBytes: newRemovableGauge(
			"speedtest_download_bytes",
			"Bytes transferred during the most recent Speedtest download phase.",
		),
		downloadElapsed: newRemovableGauge(
			"speedtest_download_elapsed_seconds",
			"Elapsed time of the most recent Speedtest download phase in seconds.",
		),
		downloadLatencyIQM: newRemovableGauge(
			"speedtest_download_latency_iqm_milliseconds",
			"Interquartile mean latency during the most recent Speedtest download phase in milliseconds.",
		),
		downloadLatencyLow: newRemovableGauge(
			"speedtest_download_latency_low_milliseconds",
			"Lowest latency during the most recent Speedtest download phase in milliseconds.",
		),
		downloadLatencyHigh: newRemovableGauge(
			"speedtest_download_latency_high_milliseconds",
			"Highest latency during the most recent Speedtest download phase in milliseconds.",
		),
		downloadLatencyJitter: newRemovableGauge(
			"speedtest_download_latency_jitter_milliseconds",
			"Jitter during the most recent Speedtest download phase in milliseconds.",
		),
		upload: newRemovableGauge(
			"speedtest_upload_bits_per_second",
			"Upload bandwidth from the most recent Speedtest measurement in bits per second.",
		),
		uploadBytes: newRemovableGauge(
			"speedtest_upload_bytes",
			"Bytes transferred during the most recent Speedtest upload phase.",
		),
		uploadElapsed: newRemovableGauge(
			"speedtest_upload_elapsed_seconds",
			"Elapsed time of the most recent Speedtest upload phase in seconds.",
		),
		uploadLatencyIQM: newRemovableGauge(
			"speedtest_upload_latency_iqm_milliseconds",
			"Interquartile mean latency during the most recent Speedtest upload phase in milliseconds.",
		),
		uploadLatencyLow: newRemovableGauge(
			"speedtest_upload_latency_low_milliseconds",
			"Lowest latency during the most recent Speedtest upload phase in milliseconds.",
		),
		uploadLatencyHigh: newRemovableGauge(
			"speedtest_upload_latency_high_milliseconds",
			"Highest latency during the most recent Speedtest upload phase in milliseconds.",
		),
		uploadLatencyJitter: newRemovableGauge(
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
		up: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "speedtest_up",
			Help: "Whether the most recent Speedtest measurement succeeded (1) or failed (0).",
		}),
		runTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "speedtest_runs_total",
			Help: "Total Speedtest CLI runs, partitioned by outcome.",
		}, []string{"result"}),
		duration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "speedtest_run_duration_seconds",
			Help:    "Wall-clock duration of a Speedtest CLI run.",
			Buckets: []float64{5, 15, 30, 60, 90, 120, 180},
		}),
		lastRun: newRemovableGauge(
			"speedtest_last_run_timestamp_seconds",
			"Unix timestamp when the most recent successful Speedtest run completed.",
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

type removableGauge struct {
	*prometheus.GaugeVec
}

func newRemovableGauge(name, help string) removableGauge {
	return removableGauge{GaugeVec: prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: name, Help: help}, nil)}
}

func (g removableGauge) Set(value float64) {
	g.WithLabelValues().Set(value)
}

func (g removableGauge) Delete() {
	g.DeleteLabelValues()
}

// Update ensures metrics contain a fresh measurement. Failures leave the
// scrape itself successful but omit invalid measurement samples and set
// speedtest_up to zero, allowing Prometheus to alert on the measurement.
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
	c.hasResult = true
	c.cacheExpires = completed.Add(c.cacheFor)

	if err != nil {
		c.setFailure()
		c.runTotal.WithLabelValues("error").Inc()
		return err
	}
	c.lastRun.Set(float64(completed.UnixNano()) / float64(time.Second))
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
	c.serverID.Delete()
	c.jitter.Delete()
	c.ping.Delete()
	c.pingLow.Delete()
	c.pingHigh.Delete()
	c.download.Delete()
	c.downloadBytes.Delete()
	c.downloadElapsed.Delete()
	c.downloadLatencyIQM.Delete()
	c.downloadLatencyLow.Delete()
	c.downloadLatencyHigh.Delete()
	c.downloadLatencyJitter.Delete()
	c.upload.Delete()
	c.uploadBytes.Delete()
	c.uploadElapsed.Delete()
	c.uploadLatencyIQM.Delete()
	c.uploadLatencyLow.Delete()
	c.uploadLatencyHigh.Delete()
	c.uploadLatencyJitter.Delete()
	c.lastRun.Delete()
	c.packetLoss.DeleteLabelValues()
	if c.currentISP != "" {
		c.ispInfo.DeleteLabelValues(c.currentISP)
		c.currentISP = ""
	}
	c.up.Set(0)
}
