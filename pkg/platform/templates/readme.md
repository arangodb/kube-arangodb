<!--
DISCLAIMER

Copyright 2025-2026 ArangoDB GmbH, Cologne, Germany

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

Copyright holder is ArangoDB GmbH, Cologne, Germany
-->

# {{ .Name }}

Arango Platform Release `{{ .Version }}`.

This chart installs an Arango Platform release onto an existing `ArangoDeployment`. It bundles
{{ len .Charts }} platform chart(s) and {{ len .Services }} service(s), which are created as
`ArangoPlatformChart` and `ArangoPlatformService` resources and reconciled by the ArangoDB operator.

> This file is generated when the release chart is packaged. Do not edit it by hand.

## Installation

The name of the target `ArangoDeployment` is required:

```bash
helm install <release-name> <chart> \
  --namespace <namespace> \
  --set deployment=<arango-deployment-name>
```

## Values

| Key | Type | Required | Description |
|-----|------|----------|-------------|
| `deployment` | string | yes | Name of the `ArangoDeployment` to target. Must be non-empty. |
| `charts.<chart>` | object | no | Value overrides for a bundled chart. |
| `services.<service>.values` | object | no | Value overrides for a service. |

Overrides are merged on top of the values each chart was packaged with, so only the keys you want
to change need to be listed. Unknown chart or service names are rejected by `values.schema.json`.

A complete `values.yaml` for this release, with every override block left empty:

```yaml
deployment: my-deployment
{{- if .Charts }}
charts:
{{- range $key, $value := .Charts }}
  {{ $value.Name }}: {}
{{- end }}
{{- end }}
{{- if .Services }}
services:
{{- range $key, $value := .Services }}
  {{ $value.Name }}:
    values: {}
{{- end }}
{{- end }}
```

## Bundled charts
{{ if .Charts }}
| Chart | Version | Override validation |
|-------|---------|---------------------|
{{- range $key, $value := .Charts }}
| `{{ $value.Name }}` | `{{ $value.Version }}` | {{ if $value.Schema }}validated against the chart's own `values.schema.json`{{ else }}not validated - chart ships no schema{{ end }} |
{{- end }}

Charts that ship a `values.schema.json` have their override block validated against it, with
`required` constraints relaxed because overrides are a partial document. Charts without a schema
accept any values.
{{ else }}
This release bundles no charts.
{{ end }}
## Services
{{ if .Services }}
| Service | Chart |
|---------|-------|
{{- range $key, $value := .Services }}
| `{{ $value.Name }}` | `{{ $value.ChartRef }}` |
{{- end }}

### Service values

Top-level values each service exposes, with the defaults it is packaged with. Set them under
`services.<service>.values`. Descriptions are taken from the chart's `values.schema.json`
where it provides them.
{{ range $key, $value := .Services }}
#### `{{ $value.Name }}`

Chart `{{ $value.ChartRef }}`.
{{ if $value.Values }}
| Value | Default | Description |
|-------|---------|-------------|
{{- range $v := $value.Values }}
| `{{ $v.Key }}` | `{{ $v.Default }}` | {{ if $v.Description }}{{ $v.Description }}{{ else }}-{{ end }} |
{{- end }}
{{ else }}
This service exposes no values.
{{ end }}{{ end }}{{ else }}
This release bundles no services.
{{ end }}
## Verifying the installation

```bash
kubectl get arangoplatformcharts -n <namespace>
kubectl get arangoplatformservices -n <namespace>
```

Both resources report a `Ready` condition once the operator has reconciled them.
