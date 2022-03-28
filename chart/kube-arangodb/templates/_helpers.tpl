{{/* vim: set filetype=mustache: */}}

{{/*
Expand the name of the chart.
*/}}
{{- define "kube-arangodb.name" -}}
{{- printf "%s" .Chart.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Expand the name of the release.
*/}}
{{- define "kube-arangodb.releaseName" -}}
{{- printf "%s" .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Expand the name of the operator.
*/}}
{{- define "kube-arangodb.operatorName" -}}
{{- printf "arango-%s-operator" .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Combine name of the deployment.
*/}}
{{- define "kube-arangodb.fullName" -}}
{{- printf "%s-%s" .Chart.Name .Release.Name  | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create the name of the Operator RBAC role
*/}}
{{- define "kube-arangodb.rbac" -}}
{{- printf "%s-%s" (include "kube-arangodb.operatorName" .) "rbac" | trunc 95 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create the name of the Operator Cluster resources
*/}}
{{- define "kube-arangodb.rbac-cluster" -}}
{{- if eq .Release.Namespace "default" -}}
{{- printf "%s-rbac" (include "kube-arangodb.operatorName" .) | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s-rbac" (include "kube-arangodb.operatorName" .) .Release.Namespace | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
