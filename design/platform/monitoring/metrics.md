# Platform Metrics Integration

## Overview

Platform Metrics provides a standardized way for pods to expose Prometheus
metrics and be automatically discovered by the platform monitoring stack.
Integration is annotation-driven: pods opt in by adding annotations, and
Prometheus discovers them via Kubernetes service discovery.

## Architecture

```
+------------------+       scrape        +---------------------+
|                  | <------------------ |                     |
|   Pod            |   (annotations)     |   Prometheus        |
|   /metrics:9102  |                     |   (k8s_sd: pod)     |
|                  |                     |                     |
+------------------+                     +---------------------+
                                                  |
                                                  | query
                                                  v
                                         +---------------------+
                                         |                     |
                                         |   Grafana           |
                                         |   /grafana/         |
                                         |                     |
                                         +---------------------+
```

Prometheus uses `kubernetes_sd_configs` with `role: pod` scoped to the release
namespace. It filters pods based on the `platform.arangodb.com/scrape`
annotation.

## Pod Annotations

All annotations use the `platform.arangodb.com` prefix.

### Required

| Annotation | Value | Description |
|---|---|---|
| `platform.arangodb.com/scrape` | `"true"` | Enables scraping for this pod |

### Optional

| Annotation | Default | Description |
|---|---|---|
| `platform.arangodb.com/port` | `9102` | Port exposing metrics |
| `platform.arangodb.com/path` | `/metrics` | HTTP path for metrics endpoint |
| `platform.arangodb.com/scheme` | `http` | Protocol (`http` or `https`) |
| `platform.arangodb.com/scrape-slow` | - | If `"true"`, scrape at 5m interval instead of default 1m. Mutually exclusive with `scrape` |
| `platform.arangodb.com/param_<name>` | - | Passed as query parameter `<name>` to the scrape request |

### Example

Minimal pod integration:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: my-service
  annotations:
    platform.arangodb.com/scrape: "true"
    platform.arangodb.com/port: "8080"
    platform.arangodb.com/path: "/metrics"
spec:
  containers:
    - name: app
      ports:
        - containerPort: 8080
```

## Operator Integration

The kube-arangodb operator automatically sets metrics annotations on ArangoDB
pods when `spec.metrics.enabled: true` is set on the `ArangoDeployment`
resource.

The operator uses constants defined in `pkg/util/constants/constants.go`:

```go
AnnotationMetricsScrapeLabel = "platform.arangodb.com/scrape"
AnnotationMetricsScrapePort  = "platform.arangodb.com/port"
```

In `pkg/deployment/resources/pod_creator_arangod.go`, the pod creator adds
these annotations when metrics are enabled:

```go
result[AnnotationMetricsScrapeLabel] = "true"
result[AnnotationMetricsScrapePort]  = fmt.Sprintf("%d", m.GroupSpec.GetExporterPort())
```

## Scrape Jobs

Prometheus is configured with two pod scrape jobs:

### `kubernetes-pods` (default)

- Scrape interval: 1m (global default)
- Selects pods with `platform.arangodb.com/scrape: "true"`
- Drops pods that also have `platform.arangodb.com/scrape-slow: "true"`
- TLS verification is skipped (`insecure_skip_verify: true`)
- Drops pods in `Pending`, `Succeeded`, `Failed`, or `Completed` phase

### `kubernetes-pods-slow`

- Scrape interval: 5m
- Scrape timeout: 30s
- Selects pods with `platform.arangodb.com/scrape-slow: "true"`
- Same relabeling rules as the default job

## Labels

Prometheus automatically adds the following labels to scraped metrics:

| Label | Source |
|---|---|
| `namespace` | Pod namespace |
| `pod` | Pod name |
| `node` | Node the pod runs on |
| All pod labels | Mapped via `labelmap` |

## Namespace Scoping

Prometheus is scoped to the Helm release namespace by default. Pods in other
namespaces are not discovered unless additional namespaces are configured in the
Prometheus chart values.

## Security Considerations

- Metrics endpoints should not expose sensitive data.
- HTTPS endpoints are supported via the `scheme` annotation. When using HTTPS,
  TLS verification is skipped by default in the scrape config.
- Pods exposing metrics should ensure the metrics port is not publicly
  accessible outside the cluster.

## Prerequisites

- `platform-monitoring-prometheus` Helm chart deployed in the target namespace.
- kube-arangodb >= 1.3.4 for operator-managed ArangoDB pods.
- `spec.metrics.enabled: true` in the `ArangoDeployment` spec for ArangoDB
  metrics.
