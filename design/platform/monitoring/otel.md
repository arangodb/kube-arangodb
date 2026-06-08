# OpenTelemetry Integration

## Overview

The OpenTelemetry (OTEL) integration exposes standard OTEL environment variables
to containers managed by the operator. OTEL SDKs read these natively — no custom
mapping or wrapper is needed.

The integration is configured entirely through `ArangoProfile` — no changes to
the ArangoDB deployment spec are needed.

## Environment Variables

The integration uses standard [OpenTelemetry environment variables](https://opentelemetry.io/docs/specs/otel/configuration/sdk-environment-variables/)
directly. OTEL SDKs auto-configure from these without any application changes.

### Exporter Configuration

| Variable | Description | Example |
|---|---|---|
| `OTEL_EXPORTER_OTLP_ENDPOINT` | OTLP collector endpoint (gRPC or HTTP) | `http://otel-collector.monitoring:4317` |
| `OTEL_EXPORTER_OTLP_PROTOCOL` | Transport protocol | `grpc`, `http/protobuf` |
| `OTEL_EXPORTER_OTLP_HEADERS` | Additional headers for the exporter | `Authorization=Bearer xxx` |
| `OTEL_EXPORTER_OTLP_CERTIFICATE` | Path to TLS CA certificate | `/certs/ca.crt` |
| `OTEL_EXPORTER_OTLP_INSECURE` | Disable TLS verification | `true` |

### Signal-specific Endpoints (optional overrides)

| Variable | Description |
|---|---|
| `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT` | Override endpoint for traces only |
| `OTEL_EXPORTER_OTLP_METRICS_ENDPOINT` | Override endpoint for metrics only |
| `OTEL_EXPORTER_OTLP_LOGS_ENDPOINT` | Override endpoint for logs only |

### Resource Attributes

| Variable | Description | Example |
|---|---|---|
| `OTEL_SERVICE_NAME` | Logical service name | `arangodb-coordinator` |
| `OTEL_RESOURCE_ATTRIBUTES` | Comma-separated key=value resource attributes | `deployment.name=mydb,k8s.namespace.name=prod` |

### SDK Configuration

| Variable | Description | Default |
|---|---|---|
| `OTEL_SDK_DISABLED` | Disable the SDK entirely | `false` |
| `OTEL_TRACES_SAMPLER` | Sampler type | `parentbased_traceidratio` |
| `OTEL_TRACES_SAMPLER_ARG` | Sampler argument (e.g. ratio) | `0.1` |
| `OTEL_LOG_LEVEL` | SDK log level | `info` |
| `OTEL_PROPAGATORS` | Context propagators | `tracecontext,baggage` |

## ArangoProfile Configuration

OTEL environment variables are set through the `ArangoProfile` CRD's container
template. The `all` field applies environment variables to every container in
the pod, including ArangoDB servers and sidecars.

### Example: Basic OTLP Export

```yaml
apiVersion: scheduler.arangodb.com/v1beta1
kind: ArangoProfile
metadata:
  name: otel-export
  namespace: default
spec:
  selectors:
    label:
      matchLabels:
        app: arangodb
  template:
    container:
      all:
        env:
          - name: OTEL_EXPORTER_OTLP_ENDPOINT
            value: "http://otel-collector.monitoring:4317"
          - name: OTEL_EXPORTER_OTLP_PROTOCOL
            value: "grpc"
          - name: OTEL_SERVICE_NAME
            value: "arangodb"
          - name: OTEL_RESOURCE_ATTRIBUTES
            value: "deployment.environment=production"
```

### Example: Traces Only with Sampling

```yaml
apiVersion: scheduler.arangodb.com/v1beta1
kind: ArangoProfile
metadata:
  name: otel-traces
  namespace: default
spec:
  selectors:
    label:
      matchLabels:
        app: arangodb
  template:
    container:
      all:
        env:
          - name: OTEL_EXPORTER_OTLP_TRACES_ENDPOINT
            value: "http://tempo.monitoring:4317"
          - name: OTEL_EXPORTER_OTLP_PROTOCOL
            value: "grpc"
          - name: OTEL_TRACES_SAMPLER
            value: "parentbased_traceidratio"
          - name: OTEL_TRACES_SAMPLER_ARG
            value: "0.01"
          - name: OTEL_SERVICE_NAME
            value: "arangodb"
```

### Example: Using a Secret for Headers

```yaml
apiVersion: scheduler.arangodb.com/v1beta1
kind: ArangoProfile
metadata:
  name: otel-auth
  namespace: default
spec:
  template:
    container:
      all:
        env:
          - name: OTEL_EXPORTER_OTLP_ENDPOINT
            value: "https://otlp.vendor.io:443"
          - name: OTEL_EXPORTER_OTLP_PROTOCOL
            value: "http/protobuf"
          - name: OTEL_EXPORTER_OTLP_HEADERS
            valueFrom:
              secretKeyRef:
                name: otel-credentials
                key: headers
```

### Example: Per-Container Override

```yaml
apiVersion: scheduler.arangodb.com/v1beta1
kind: ArangoProfile
metadata:
  name: otel-per-container
spec:
  template:
    container:
      all:
        env:
          - name: OTEL_EXPORTER_OTLP_ENDPOINT
            value: "http://otel-collector.monitoring:4317"
      containers:
        server:
          env:
            - name: OTEL_SERVICE_NAME
              value: "arangodb-server"
        sidecar:
          env:
            - name: OTEL_SERVICE_NAME
              value: "arangodb-sidecar"
```

## How It Works

1. The `ArangoProfile` is matched to pods via label selectors
2. The profile template's `container.all.env` entries are merged into every
   container in the pod template
3. Specific containers can be targeted via `container.containers.<name>.env`
4. Environment variables support standard Kubernetes `EnvVar` — plain values,
   `valueFrom.secretKeyRef`, `valueFrom.configMapKeyRef`, and
   `valueFrom.fieldRef`
5. OTEL SDKs auto-configure from these variables — no application changes needed

The operator enriches `OTEL_RESOURCE_ATTRIBUTES` with deployment-level metadata
if not already present:

- `k8s.namespace.name` — from pod namespace
- `k8s.pod.name` — from pod name
- `arangodb.deployment.name` — from deployment name

## Scope

This integration covers **environment variable injection only**. It does not:

- Deploy an OTEL collector (use a separate Helm chart or operator)
- Auto-instrument applications beyond what the SDK does with env vars
- Manage TLS certificates for the exporter

## Key Files

- `pkg/apis/scheduler/v1beta1/profile_spec.go` — ArangoProfile spec
- `pkg/apis/scheduler/v1beta1/profile_template.go` — Template with pod/container fields
- `pkg/apis/scheduler/v1beta1/profile_container_template.go` — Container template (all, containers, default)
- `pkg/apis/scheduler/v1beta1/container/generic.go` — Generic container with Environments
