# speedtest-exporter

Prometheus exporter backed by the official Ookla Speedtest CLI

**Homepage:** <https://github.com/ishioni/speedtest-exporter>

> A Speedtest run consumes real WAN bandwidth. Set `config.cacheFor` to at least
> the scrape interval; each replica runs tests independently.

## Install

```sh
helm upgrade --install speedtest-exporter \
  oci://ghcr.io/ishioni/charts/speedtest-exporter \
  --namespace monitoring --create-namespace \
  --set monitoring.serviceMonitor.enabled=true
```

## Source Code

* <https://github.com/ishioni/speedtest-exporter>

## Requirements

Kubernetes: `>=1.25.0-0`

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | Affinity rules for the exporter Pod. |
| config.cacheFor | string | `"30m"` | How long to reuse the latest completed result. `0` runs a test on every scrape. |
| config.extraEnv | list | `[]` | Extra environment variables merged into the exporter container. |
| config.logFormat | string | `"json"` | Log format: json or text. |
| config.logLevel | string | `"info"` | Log level: debug, info, warn, or error. |
| config.serverID | string | `""` |  |
| config.timeout | string | `"90s"` | Maximum duration for a CLI run in Go duration syntax, for example 90s or 2m. |
| deploymentAnnotations | object | `{}` | Deployment annotations. |
| fullnameOverride | string | `""` | Override the full release name. |
| image.digest | string | `""` | Image digest (sha256:…); when set, it overrides tag. |
| image.pullPolicy | string | `"IfNotPresent"` | Image pull policy. |
| image.repository | string | `"ghcr.io/ishioni/speedtest-exporter"` | Image repository. |
| image.tag | string | `""` | Image tag; defaults to the chart appVersion. |
| imagePullSecrets | list | `[]` | Image pull secrets for private registries. |
| livenessProbe | object | `{"httpGet":{"path":"/healthz","port":"http"},"initialDelaySeconds":5,"periodSeconds":20}` | Liveness probe; it verifies that the HTTP server is responsive, not WAN connectivity. |
| monitoring.dashboards.annotations | object | `{}` | Annotations added to the dashboard ConfigMap. |
| monitoring.dashboards.enabled | bool | `false` | Render the Grafana dashboard ConfigMap for the Grafana sidecar or grafana-operator. |
| monitoring.dashboards.grafanaOperator.allowCrossNamespaceImport | bool | `true` | Allow a Grafana instance in a different namespace to import the dashboard. |
| monitoring.dashboards.grafanaOperator.enabled | bool | `false` | Render a GrafanaDashboard CR that imports the ConfigMap instead of using the Grafana sidecar label. |
| monitoring.dashboards.grafanaOperator.folder | string | `""` | Grafana folder for the dashboard. |
| monitoring.dashboards.grafanaOperator.matchLabels | object | `{}` | Labels that select the target Grafana instance. Required when grafanaOperator.enabled is true. |
| monitoring.dashboards.grafanaOperator.resyncPeriod | string | `"10m"` | How often the Grafana Operator resyncs this dashboard. |
| monitoring.dashboards.labels | object | `{}` | Labels added to the dashboard ConfigMap. For the Grafana sidecar, add any labels your deployment selects on. |
| monitoring.dashboards.namespace | string | `""` | Namespace for dashboard objects; defaults to the release namespace. |
| monitoring.prometheusRule.alerts.exporterAbsent.enabled | bool | `true` | Alert when the exporter is absent or has no healthy scrape target. |
| monitoring.prometheusRule.alerts.exporterAbsent.for | string | `"5m"` | Time the exporter must be unavailable before alerting. |
| monitoring.prometheusRule.alerts.exporterAbsent.severity | string | `"critical"` | Severity for the exporter-absent alert. |
| monitoring.prometheusRule.alerts.highJitterLatency.enabled | bool | `true` | Alert when average jitter exceeds the configured millisecond threshold. |
| monitoring.prometheusRule.alerts.highJitterLatency.for | string | `"10m"` | Time the average must remain above threshold before alerting. |
| monitoring.prometheusRule.alerts.highJitterLatency.severity | string | `"warning"` | Severity for the high-jitter alert. |
| monitoring.prometheusRule.alerts.highJitterLatency.threshold | int | `30` | Jitter threshold in milliseconds. |
| monitoring.prometheusRule.alerts.highJitterLatency.window | string | `"3h"` | Lookback window used for the average. |
| monitoring.prometheusRule.alerts.highPingLatency.enabled | bool | `true` | Alert when average ping latency exceeds the configured millisecond threshold. |
| monitoring.prometheusRule.alerts.highPingLatency.for | string | `"10m"` | Time the average must remain above threshold before alerting. |
| monitoring.prometheusRule.alerts.highPingLatency.severity | string | `"warning"` | Severity for the high-ping alert. |
| monitoring.prometheusRule.alerts.highPingLatency.threshold | int | `15` | Ping threshold in milliseconds. |
| monitoring.prometheusRule.alerts.highPingLatency.window | string | `"3h"` | Lookback window used for the average. |
| monitoring.prometheusRule.alerts.slowDownload.enabled | bool | `true` | Alert when average download speed is below the configured Mbps threshold. |
| monitoring.prometheusRule.alerts.slowDownload.for | string | `"10m"` | Time the average must remain below threshold before alerting. |
| monitoring.prometheusRule.alerts.slowDownload.severity | string | `"warning"` | Severity for the slow-download alert. |
| monitoring.prometheusRule.alerts.slowDownload.threshold | int | `100` | Download threshold in Mbps. |
| monitoring.prometheusRule.alerts.slowDownload.window | string | `"3h"` | Lookback window used for the average. |
| monitoring.prometheusRule.alerts.slowUpload.enabled | bool | `true` | Alert when average upload speed is below the configured Mbps threshold. |
| monitoring.prometheusRule.alerts.slowUpload.for | string | `"10m"` | Time the average must remain below threshold before alerting. |
| monitoring.prometheusRule.alerts.slowUpload.severity | string | `"warning"` | Severity for the slow-upload alert. |
| monitoring.prometheusRule.alerts.slowUpload.threshold | int | `10` | Upload threshold in Mbps. |
| monitoring.prometheusRule.alerts.slowUpload.window | string | `"3h"` | Lookback window used for the average. |
| monitoring.prometheusRule.annotations | object | `{}` | Annotations added to the PrometheusRule. |
| monitoring.prometheusRule.enabled | bool | `false` | Create a Prometheus Operator PrometheusRule containing availability and connection-quality alerts. |
| monitoring.prometheusRule.groupName | string | `"speedtest-exporter"` | Alert group name. |
| monitoring.prometheusRule.labels | object | `{}` | Labels added to the PrometheusRule. Set labels required by your PrometheusRule selector here. |
| monitoring.serviceMonitor.annotations | object | `{}` | ServiceMonitor annotations. |
| monitoring.serviceMonitor.enabled | bool | `false` | Create a Prometheus Operator ServiceMonitor. |
| monitoring.serviceMonitor.interval | string | `"1m"` | Scrape interval. Keep it no shorter than config.cacheFor to avoid unnecessary tests. |
| monitoring.serviceMonitor.jobLabel | string | `"app.kubernetes.io/instance"` | Service label used as Prometheus's stable job label. The default is the Helm release name. |
| monitoring.serviceMonitor.labels | object | `{}` | ServiceMonitor labels. |
| monitoring.serviceMonitor.metricRelabelings | list | `[]` | Prometheus metric relabelings. |
| monitoring.serviceMonitor.path | string | `"/metrics"` | Metrics path. |
| monitoring.serviceMonitor.relabelings | list | `[]` | Prometheus relabelings. |
| monitoring.serviceMonitor.scrapeTimeout | string | `"30s"` | Must exceed config.timeout plus CLI startup overhead. |
| nameOverride | string | `""` | Override the chart name used in resource names. |
| nodeSelector | object | `{}` | Node selector for the exporter Pod. |
| podAnnotations | object | `{}` | Pod annotations. |
| podDisruptionBudget | object | `{"enabled":false,"minAvailable":1}` | Optional PodDisruptionBudget configuration. |
| podLabels | object | `{}` | Pod labels. |
| podSecurityContext | object | `{"fsGroup":65532,"fsGroupChangePolicy":"OnRootMismatch","runAsGroup":65532,"runAsNonRoot":true,"runAsUser":65532,"seccompProfile":{"type":"RuntimeDefault"}}` | Runs the distroless non-root image with the RuntimeDefault seccomp profile. |
| readinessProbe | object | `{"httpGet":{"path":"/readyz","port":"http"},"initialDelaySeconds":5,"periodSeconds":10}` | Readiness probe; it is available after startup CLI validation. |
| replicaCount | int | `1` | Number of exporter Pods. Each replica independently performs Speedtests. |
| resources | object | `{"limits":{"memory":"128Mi"},"requests":{"cpu":"100m","memory":"64Mi"}}` | Resource requests and limits. Speed tests are network and CPU intensive. |
| securityContext | object | `{"allowPrivilegeEscalation":false,"capabilities":{"drop":["ALL"]},"readOnlyRootFilesystem":true}` | Drops all Linux capabilities and keeps the image filesystem read-only. |
| service.annotations | object | `{}` | Service annotations. |
| service.port | int | `9798` | Prometheus and health endpoint port. |
| service.type | string | `"ClusterIP"` | Service type. |
| serviceAccount.annotations | object | `{}` | ServiceAccount annotations. |
| serviceAccount.automount | bool | `false` | Automount the ServiceAccount token. |
| serviceAccount.create | bool | `true` | Create a ServiceAccount. The exporter does not need a Kubernetes API token. |
| serviceAccount.name | string | `""` | ServiceAccount name; generated if empty. |
| startupProbe | object | `{}` | Optional startup probe. Empty disables it. |
| terminationGracePeriodSeconds | int | `30` | Grace period used for in-flight HTTP requests on termination. |
| tests | object | `{"image":{"pullPolicy":"IfNotPresent","repository":"curlimages/curl","tag":"8.21.0"}}` | Image and pull policy used by `helm test`. |
| tolerations | list | `[]` | Tolerations for the exporter Pod. |

---

_This README is generated by [helm-docs](https://github.com/norwoodj/helm-docs) from `Chart.yaml` and `values.yaml`. Edit those (or `README.md.gotmpl`) and run `mise run generate`._
