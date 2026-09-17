---
layout: page
has_children: false
title: Developer Guide
parent: ArangoDBPlatform
nav_order: 5
---

# Platform Developer Guide

This guide documents behavior that authors of platform service charts (charts installed via
`ArangoPlatformService` and `ArangoPlatformWorkflow`) need to be aware of.

## Injected values

When the operator installs or upgrades a platform chart, it merges a set of operator-managed values on top
of the user-provided `spec.values`. These are exposed under the top-level `arangodb_platform` key:

```yaml
arangodb_platform:
  deployment:
    name: <ArangoDeployment name>   # always injected
  hibernated: true                  # injected only when spec.hibernate is true
```

A chart should treat `arangodb_platform` as read-only operator input and must not require the user to set
it.

## Hibernation

Both `ArangoPlatformService` and `ArangoPlatformWorkflow` expose a `spec.hibernate` flag:

```yaml
apiVersion: platform.arangodb.com/v1beta1
kind: ArangoPlatformService
metadata:
  name: my-service
spec:
  deployment:
    name: my-deployment
  chart:
    name: my-chart
  hibernate: true
```

When `spec.hibernate` is `true`, the operator injects `arangodb_platform.hibernated: true` into the chart
values. When it is `false` or unset, the `hibernated` key is omitted entirely (charts must not rely on it
being present).

Changing `spec.hibernate` re-renders the release with the new value, so the chart is upgraded in place (no
resource is deleted); the deployment name is preserved.

### Handling hibernation in a chart

A chart reacts to the flag by rendering its hibernated form when the value is set. For example, scaling a
workload down while hibernated:

```yaml
spec:
  replicas: {{ if .Values.arangodb_platform.hibernated }}0{{ else }}{{ .Values.replicas }}{{ end }}
```

or guarding entire resources:

```yaml
{{- if not .Values.arangodb_platform.hibernated }}
# resources that should not run while hibernated (e.g. CronJobs, workers)
{{- end }}
```

Because `hibernated` is omitted when `false`, always access it defensively, e.g.
`{{ if .Values.arangodb_platform.hibernated }}`; do not assume the key exists.

Hibernation only changes the values passed to the chart — the operator does not itself stop, delete, or
scale any workload. What "hibernated" means for a given service is entirely defined by how that service's
chart renders the flag.
