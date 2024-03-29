{{- define "exporter.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "exporter.fullname" -}}
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
Create chart name and version as used by the chart label.
*/}}
{{- define "exporter.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "exporter.labels" -}}
helm.sh/chart: {{ include "exporter.chart" . }}
{{ include "exporter.selectorLabels" . }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "exporter.selectorLabels" -}}
app.kubernetes.io/name: {{ include "exporter.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{- define "parse.resource.string" }}
  {{- $arg := . }}
	{{- $splitString := "/" }}
	{{- $resourceSplitted := split $splitString $arg }}
	{{- if eq ( len $resourceSplitted ) 2 }}
		{{- dict "resource" $resourceSplitted._0 "api" $resourceSplitted._1 | toJson }}
	{{- else if eq ( len $resourceSplitted ) 3 }}
	  {{- dict "resource" $resourceSplitted._0 "version" $resourceSplitted._1 "api" $resourceSplitted._2 | toJson }}
	{{- else }}
    {{- fail ( printf "Can't parse resource string %s" $arg ) }}
  {{- end }}
{{- end }}

{{- define "format.prom.label" }}
	{{- $arg := . }}
	{{- $result := replace "/" "_" $arg }}
	{{- $result = replace "." "_" $result }}
	{{- $result = replace "-" "_" $result }}
	{{- lower $result }}
{{- end }}
