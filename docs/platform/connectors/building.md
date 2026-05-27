---
layout: page
title: Building a Connector
parent: Connectors
grand_parent: ArangoDBPlatform
nav_order: 4
---

# Building a Connector

A connector is a program that communicates with the integration sidecar's
internal API to process jobs.

## Requirements

- Runs as a container in a pod with integration sidecar labels
- Communicates via HTTP with the internal gateway (`http://127.0.0.1:9192`)
- Polls for jobs, executes them, uploads results

## Step 1: Create the ArangoPlatformConnector CRD

Define what your connector does and what input it accepts:

```yaml
apiVersion: networking.arangodb.com/v1beta1
kind: ArangoRoute
metadata:
  name: my-connector-route
spec:
  deployment: <deployment-name>
  route:
    path: /connector/my-connector/
  destination:
    path: /_integration/connector/v1/
---
apiVersion: platform.arangodb.com/v1beta1
kind: ArangoPlatformConnector
metadata:
  name: my-connector
spec:
  type: Active
  deployment:
    name: <deployment-name>
  route:
    name: my-connector-route
  description: "What this connector does"
  tags:
    - my-tag
  schema:
    type: object
    properties:
      myParam:
        type: string
    required:
      - myParam
  version: "1.0.0"
```

The `ArangoRoute` redirects `/connector/my-connector/*` to the internal
`/_integration/connector/v1/*` endpoint, giving AI tools a clean URL.

## Step 2: Implement the Connector Loop

Your connector binary should:

```
loop:
  POST /_internal/connector/v1/job/pickup
  if no job → sleep and retry

  GET /_internal/connector/v1/job/{id}
  parse job.query

  POST /_internal/connector/v1/job/{id}/status  → Running

  execute the work...

  POST /_internal/connector/v1/job/{id}/upload/{filename}
  (upload result files)

  POST /_internal/connector/v1/job/{id}/status  → Completed (or Failed)
```

## Step 3: Create a Helm Chart

Your chart should include:
- The `ArangoPlatformConnector` CRD
- A `Deployment` with integration sidecar labels
- A `platform.yaml` with connector metadata

### Deployment Labels

```yaml
metadata:
  labels:
    profiles.arangodb.com/apply: "yes"
    profiles.arangodb.com/deployment: {{ .Values.arangodb_platform.deployment.name }}
```

These labels enable the integration sidecar injection.

### Standard Values

Accept the standard platform input:

```yaml
# values.yaml
arangodb_platform:
  deployment:
    name: ""
```

This is automatically populated when deploying via `ArangoPlatformService`.

## Step 4: Deploy

```bash
helm install my-connector ./chart \
  --namespace <namespace> \
  --set arangodb_platform.deployment.name=<deployment-name>
```

Or deploy via `ArangoPlatformService` for managed lifecycle.

## Example

See the sample AQL connector in `modules/test/tests/connectors/aql/` for a
complete working example.
