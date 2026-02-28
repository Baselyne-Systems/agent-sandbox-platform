{{/*
Expand the name of the chart.
*/}}
{{- define "bulkhead.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "bulkhead.fullname" -}}
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

{{/*
Common labels.
*/}}
{{- define "bulkhead.labels" -}}
app.kubernetes.io/name: {{ include "bulkhead.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version }}
{{- end }}

{{/*
Selector labels.
*/}}
{{- define "bulkhead.selectorLabels" -}}
app.kubernetes.io/name: {{ include "bulkhead.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Database URL.
*/}}
{{- define "bulkhead.databaseUrl" -}}
postgres://{{ .Values.postgres.credentials.username }}:{{ .Values.postgres.credentials.password }}@{{ include "bulkhead.fullname" . }}-postgres:5432/{{ .Values.postgres.credentials.database }}?sslmode=disable
{{- end }}
