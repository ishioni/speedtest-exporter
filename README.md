# speedtest-exporter

[![Tests](https://github.com/ishioni/speedtest-exporter/actions/workflows/tests.yaml/badge.svg)](https://github.com/ishioni/speedtest-exporter/actions/workflows/tests.yaml)
[![Lint](https://github.com/ishioni/speedtest-exporter/actions/workflows/lint.yaml/badge.svg)](https://github.com/ishioni/speedtest-exporter/actions/workflows/lint.yaml)

A Prometheus exporter that runs the **official [Ookla Speedtest CLI](https://www.speedtest.net/apps/cli)** when Prometheus scrapes it. The exporter is written in Go; the Ookla CLI remains the sole test engine.

> **Bandwidth warning:** every uncached scrape performs a real download and upload test. Configure `SPEEDTEST_CACHE_FOR` to be at least your Prometheus scrape interval. Do not deploy multiple replicas unless you want each replica to run independent tests.

## Quick start

Run the published image:

```sh
docker run --rm -p 9798:9798 \
  -e SPEEDTEST_CACHE_FOR=5m \
  ghcr.io/ishioni/speedtest-exporter:latest
```

Or run locally after installing the official `speedtest` executable on your `PATH`:

```sh
mise install
go run ./cmd/speedtest-exporter
curl http://localhost:9798/metrics
```

The process verifies `speedtest --version` at startup and refuses a non-Ookla executable.

## Prometheus

Scrape `http://<target>:9798/metrics`. The first uncached scrape runs the CLI synchronously, so the Prometheus `scrape_timeout` must exceed `SPEEDTEST_TIMEOUT` (90 seconds by default).

```yaml
scrape_configs:
  - job_name: speedtest
    scrape_interval: 5m
    scrape_timeout: 120s
    static_configs:
      - targets: [speedtest-exporter:9798]
```

### Metrics

The rewrite retains the established application metric names:

| Metric                                                                                            | Meaning                                                                                                     |
| ------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------- |
| `speedtest_server_id`                                                                             | Ookla server ID used by the latest test.                                                                    |
| `speedtest_isp_info{isp="…"}`                                                                     | ISP reported by the most recent successful test. The exporter removes the prior ISP series when it changes. |
| `speedtest_jitter_latency_milliseconds`                                                           | Latest unloaded ping jitter in milliseconds.                                                                |
| `speedtest_ping_latency_milliseconds`                                                             | Latest unloaded ping latency in milliseconds.                                                               |
| `speedtest_ping_low_latency_milliseconds` / `speedtest_ping_high_latency_milliseconds`            | Low/high unloaded ping latency range in milliseconds.                                                       |
| `speedtest_download_bits_per_second` / `speedtest_upload_bits_per_second`                         | Latest throughput in bit/s.                                                                                 |
| `speedtest_download_bytes` / `speedtest_upload_bytes`                                             | Bytes transferred during the latest download/upload phase; gauges, not counters.                            |
| `speedtest_download_elapsed_seconds` / `speedtest_upload_elapsed_seconds`                         | Duration of the latest transfer phases.                                                                     |
| `speedtest_download_latency_iqm_milliseconds` / `speedtest_upload_latency_iqm_milliseconds`       | Interquartile mean latency under download/upload load; useful bufferbloat signals.                          |
| `speedtest_download_latency_low_milliseconds` / `speedtest_upload_latency_low_milliseconds`       | Low latency under transfer load.                                                                            |
| `speedtest_download_latency_high_milliseconds` / `speedtest_upload_latency_high_milliseconds`     | High latency under transfer load.                                                                           |
| `speedtest_download_latency_jitter_milliseconds` / `speedtest_upload_latency_jitter_milliseconds` | Jitter under transfer load.                                                                                 |
| `speedtest_packet_loss_percent`                                                                   | Packet loss reported by Ookla, in percent. Omitted when the CLI does not provide it.                        |
| `speedtest_up`                                                                                    | `1` after a successful test, otherwise `0`.                                                                 |
| `speedtest_runs_total`                                                                            | CLI runs partitioned by `result` (`success` or `error`).                                                    |
| `speedtest_run_duration_seconds`                                                                  | CLI wall-clock duration histogram.                                                                          |
| `speedtest_last_run_timestamp_seconds`                                                            | Completion time of the latest CLI run.                                                                      |

Go and process collector metrics are also exposed. A CLI error still returns HTTP 200 with `speedtest_up 0`; alert on that metric to distinguish an unsuccessful measurement from an unreachable exporter.

## Configuration

All configuration is environment based:

| Variable                   | Default           | Description                                                                 |
| -------------------------- | ----------------- | --------------------------------------------------------------------------- |
| `SPEEDTEST_PORT`           | `9798`            | Legacy-compatible listen port.                                              |
| `SPEEDTEST_LISTEN_ADDRESS` | `:SPEEDTEST_PORT` | Full listen address; overrides the port variable.                           |
| `SPEEDTEST_BINARY`         | `speedtest`       | Path/name of the official Ookla executable.                                 |
| `SPEEDTEST_SERVER`         | unset             | Optional Ookla server ID.                                                   |
| `SPEEDTEST_TIMEOUT`        | `90`              | CLI time limit. Accepts legacy whole seconds or a Go duration such as `2m`. |
| `SPEEDTEST_CACHE_FOR`      | `0`               | Result cache TTL. `0` runs a test for every scrape.                         |
| `SPEEDTEST_LOG_LEVEL`      | `info`            | `debug`, `info`, `warn`, or `error`.                                        |
| `SPEEDTEST_LOG_FORMAT`     | `json`            | `json` or `text`.                                                           |

Endpoints are `/metrics`, `/healthz`, `/readyz`, and `/`.

## Kubernetes / Helm

The chart has secure defaults (non-root, read-only root filesystem, dropped Linux capabilities) and provides an optional Prometheus Operator `ServiceMonitor`:

```sh
helm upgrade --install speedtest-exporter ./charts/speedtest-exporter \
  --namespace monitoring --create-namespace \
  --set monitoring.serviceMonitor.enabled=true \
  --set config.cacheFor=5m
```

See [`charts/speedtest-exporter/values.yaml`](charts/speedtest-exporter/values.yaml) for all values. The chart intentionally defaults to a five-minute cache and a 120-second ServiceMonitor timeout.

Set `monitoring.dashboards.enabled=true` to ship a Grafana dashboard as a sidecar-discovered ConfigMap. For the Grafana Operator, also set `monitoring.dashboards.grafanaOperator.enabled=true` and provide `monitoring.dashboards.grafanaOperator.matchLabels` for the target Grafana instance; the chart then renders a `GrafanaDashboard` CR referencing the same ConfigMap.

## Local observability harness

Run the current branch against local Prometheus and Grafana—without publishing an image or deploying a chart:

```sh
mise run dev-observability-up
```

Grafana is available at <http://localhost:3000> (`admin` / `admin`); Prometheus is at <http://localhost:9090>. The harness provisions the dashboard from the exact JSON file shipped in the Helm chart. See [`deploy/observability/README.md`](deploy/observability/README.md) for lifecycle commands and the bandwidth implications of the first live test.

## Development

[Mise](https://mise.jdx.dev/) pins the project toolchain and exposes common commands:

```sh
mise install
mise run fmt
mise run test
mise run lint
mise run helm-lint
mise run helm-template
mise run dev-observability-up
```

Install the local Git hooks with `lefthook install` (or let `mise` run the post-install hook). The hooks format staged Go files, refresh generated chart artifacts, and run tests before pushing.

## License

This project is licensed under the [GNU GPL v3.0](LICENSE).
