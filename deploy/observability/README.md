# Local Prometheus and Grafana harness

This Docker Compose stack builds the exporter from the current working tree,
scrapes it with Prometheus, and provisions Grafana with the same dashboard JSON
that the Helm chart distributes. No image push, release, Kubernetes cluster, or
Grafana import is required.

> A first Prometheus scrape runs a real Ookla Speedtest. Prometheus schedules
> that first scrape within two minutes; the exporter caches each completed result
> for five minutes, so dashboard refreshes do not initiate extra tests.

## Start

From the repository root:

```sh
mise run dev-observability-up
```

The equivalent raw Docker command is:

```sh
docker compose -f deploy/observability/compose.yaml up --build --detach
```

Open:

| Service          | Address                                   |
| ---------------- | ----------------------------------------- |
| Grafana          | http://localhost:3000 — `admin` / `admin` |
| Prometheus       | http://localhost:9090                     |
| Exporter metrics | http://localhost:9799/metrics             |

In Grafana, open **Dashboards → Speedtest → Speedtest Exporter**. The initial
measurement can start within two minutes and take up to the configured
Prometheus scrape timeout (120 seconds). Confirm the target is healthy in the Prometheus [Targets page](http://localhost:9090/targets).

## Useful commands

```sh
# Follow all services; the exporter logs each CLI failure or successful startup.
mise run dev-observability-logs

# Stop containers but retain Grafana and Prometheus history.
mise run dev-observability-down

# Stop containers and delete local metrics/dashboard state.
mise run dev-observability-reset
```

The Compose stack uses host port `9799` for the exporter so it can run alongside
a manually started exporter on port `9798`.
