{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "geth-public.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "geth-public.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "geth-public.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create hostname
*/}}
{{- define "geth-public.node.hostname" -}}
{{- printf "%s.%s.%s" .Values.node.node.id .Values.organization.id .Values.cluster.domain | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Calculate specific storageClass value based on cloud provider.
*/}}
{{- define "geth-public.storageClass" -}}
{{- if hasPrefix "gcp-" .Values.cluster.name -}}
{{- printf "standard" -}}
{{- else -}}
{{- printf "default" -}}
{{- end -}}
{{- end -}}
