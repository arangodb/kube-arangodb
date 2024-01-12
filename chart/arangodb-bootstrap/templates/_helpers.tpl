{{/* vim: set filetype=mustache: */}}

{{/*
Expand the name of the chart.
*/}}
{{- define "arangodb-bootstrap.name" -}}
{{- printf "%s" .Chart.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Expand the name of the release.
*/}}
{{- define "arangodb-bootstrap.releaseName" -}}
{{- printf "%s" .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Combine name of the deployment.
*/}}
{{- define "arangodb-bootstrap.fullName" -}}
{{- printf "%s-%s" .Chart.Name .Release.Name  | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Get Secret Name
*/}}
{{- define "secret.name" -}}
{{- printf "PASSWORD_%s" (. | sha256sum | upper) -}}
{{- end -}}
