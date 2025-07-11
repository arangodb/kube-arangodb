---
layout: page
has_children: true
title: Integration Sidecars
parent: ArangoDBPlatform
nav_order: 2
---

# Integration 

## Sidecar

### Resource Types

Integration Sidecar is supported in a few resources managed by Operator:

- ArangoSchedulerDeployment (scheduler.arangodb.com/v1beta1)
- ArangoSchedulerBatchJob (scheduler.arangodb.com/v1beta1)
- ArangoSchedulerCronJob (scheduler.arangodb.com/v1beta1)
- ArangoSchedulerPod (scheduler.arangodb.com/v1beta1)

Standard Kubernetes Resources (like Pod) are also supported with Webhook extension is enabled.

To enable integration sidecar for specific deployment label needs to be defined:

```yaml
metadata:
  labels:
    profiles.arangodb.com/deployment: << deployment name >>
```

### Webhooks

When Webhook support is enabled Integration Sidecar is supported in Kubernetes Pod resources.

To inject integration sidecar for specific deployment label needs to be defined:

```yaml
apiVersion: v1
kind: Pod
metadata:
  labels:
    profiles.arangodb.com/deployment: << deployment name >>
```

### Integrations

To enable integration in specific version, labels needs to be added:

```yaml
metadata:
  labels:
    integration.profiles.arangodb.com/<< integration name >>: << integration version >>
```

#### [Authentication V1](./integration/authentication.v1.md)

Authentication Integration Sidecar

To enable:

```yaml
metadata:
  labels:
    integration.profiles.arangodb.com/authn: v1
```

#### [Authorization V0](./integration/authorization.v0.md)

Authorization Integration Sidecar

To enable:

```yaml
metadata:
  labels:
    integration.profiles.arangodb.com/authz: v0
```

#### [Scheduler V2](./integration/scheduler.v2.md)

Scheduler Integration Sidecar

To enable:

```yaml
metadata:
  labels:
    integration.profiles.arangodb.com/sched: v2
```

#### [Storage V1](./integration/storage.v1.md)

Storage Integration Sidecar (legacy)

To enable:

```yaml
metadata:
  labels:
    integration.profiles.arangodb.com/storage: v1
    
    # Optionally, enable DebugPackage integration
    storage.integration.profiles.arangodb.com/debug: "enabled"
```

#### [Storage V2](./integration/storage.v2.md)

Storage Integration Sidecar

To enable:

```yaml
metadata:
  labels:
    integration.profiles.arangodb.com/storage: v2
```

#### [Shutdown V1](./integration/shutdown.v1.md)

Shutdown Integration Sidecar

To enable:

```yaml
metadata:
  labels:
    integration.profiles.arangodb.com/shutdown: v1
```

#### [Meta V1](./integration/meta.v1.md)

Meta Integration Sidecar

To enable:

```yaml
metadata:
  labels:
    integration.profiles.arangodb.com/meta: v1
```

### Envs

#### INTEGRATION_API_ADDRESS

Integration Service API Address

Example: `localhost:1234`

#### INTEGRATION_SERVICE_ADDRESS

Integration Service API Address

Example: `localhost:1234`

#### INTEGRATION_HTTP_ADDRESS

Integration Service HTTP Address

Example: `localhost:1234`

#### ARANGO_DEPLOYMENT_NAME

ArangoDeployment name.

Example: `deployment`

#### ARANGO_DEPLOYMENT_ENDPOINT / ARANGODB_ENDPOINT

HTTP/S Endpoint of the ArangoDeployment Internal Service.

Example: `https://deployment.default.svc:8529`

#### ARANGO_DEPLOYMENT_CA (optional)

Path to the CA in the PEM format. If not set, TLS is disabled.

Example: `/etc/deployment-int/ca/ca.pem`

#### KUBERNETES_NAMESPACE

Kubernetes Namespace.

Example: `default`

#### KUBERNETES_POD_NAME

Kubernetes Pod Name.

Example: `example`

#### KUBERNETES_POD_IP

Kubernetes Pod IP.

Example: `127.0.0.1`

#### KUBERNETES_SERVICE_ACCOUNT

Kubernetes Service Account mounted for the Pod.

Example: `sa-example`

#### CONTAINER_CPU_REQUESTS

Kubernetes Pod Container CPU Requests (1000 = 1CPU).

Example: `500`

#### CONTAINER_MEMORY_REQUESTS

Kubernetes Pod Container Memory Requests in Megabytes.

Example: `128`

#### CONTAINER_CPU_LIMITS

Kubernetes Pod Container CPU Limits (1000 = 1CPU).

Example: `500`

#### CONTAINER_MEMORY_LIMITS

Kubernetes Pod Container Memory Limits in Megabytes.

Example: `128`

