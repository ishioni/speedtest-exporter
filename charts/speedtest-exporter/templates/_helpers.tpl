{{/* Expand the name of the chart. */}}
{{- define "speedtest-exporter.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/* Create a DNS-safe resource name. */}}
{{- define "speedtest-exporter.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{- define "speedtest-exporter.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "speedtest-exporter.labels" -}}
helm.sh/chart: {{ include "speedtest-exporter.chart" . }}
{{ include "speedtest-exporter.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{- define "speedtest-exporter.selectorLabels" -}}
app.kubernetes.io/name: {{ include "speedtest-exporter.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{- define "speedtest-exporter.serviceAccountName" -}}
{{- $name := tpl (.Values.serviceAccount.name | default "") $ -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "speedtest-exporter.fullname" .) $name }}
{{- else }}
{{- default "default" $name }}
{{- end }}
{{- end }}

{{/* A digest wins over a tag, allowing released charts to be immutable. */}}
{{- define "speedtest-exporter.image" -}}
{{- if .Values.image.digest -}}
{{- printf "%s@%s" .Values.image.repository .Values.image.digest -}}
{{- else -}}
{{- printf "%s:%s" .Values.image.repository (.Values.image.tag | default .Chart.AppVersion) -}}
{{- end -}}
{{- end }}
