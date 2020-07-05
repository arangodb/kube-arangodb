{{/* vim: set filetype=mustache: */}}

{{/*
Expand the name of the chart.
*/}}
{{- define "kube-arangodb-crd.name" -}}
{{- printf "%s" .Chart.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Version of apiextensions.k8s.io
*/}}
{{- define "version.apiextensions.k8s.io" -}}
{{- if and (eq .Capabilities.KubeVersion.Major "1") (gt .Capabilities.KubeVersion.Minor "15") -}}
{{- printf "v1" -}}
{{- else -}}
{{- printf "v1beta1" -}}
{{- end -}}
{{- end -}}