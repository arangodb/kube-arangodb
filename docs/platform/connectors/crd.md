---
layout: page
title: Connector CRD
parent: Connectors
grand_parent: ArangoDBPlatform
nav_order: 1
---

# ArangoPlatformConnector

The `ArangoPlatformConnector` CRD registers a connector with the platform and
makes it discoverable by AI tools via `/_inventory`.

## Example

```yaml
apiVersion: platform.arangodb.com/v1beta1
kind: ArangoPlatformConnector
metadata:
  name: aql-connector
  namespace: arangodb
spec:
  type: Active
  deployment:
    name: my-deployment
  route:
    name: aql-connector-route
  description: "Execute AQL queries on ArangoDB"
  tags:
    - database
    - aql
    - query
  schema:
    type: object
    properties:
      query:
        type: string
        description: "AQL query string"
      bindVars:
        type: object
        description: "Bind variables for the query"
    required:
      - query
  version: "1.0.0"
```

## Spec

| Field | Type | Default | Description |
|---|---|---|---|
| `type` | `string` | `Active` | Connector pattern type. Currently only `Active` is supported |
| `deployment` | `Object` | | Reference to the ArangoDeployment this connector belongs to |
| `route` | `Object` | | Reference to the ArangoRoute that exposes this connector |
| `description` | `string` | | Human-readable description of what the connector does |
| `tags` | `[]string` | | Labels for discovery and filtering |
| `schema` | `JSONSchemaProps` | | [JSON Schema](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/#specifying-a-structural-schema) defining the connector's query input parameters |
| `version` | `string` | | Connector version |

## Status

| Condition | Description |
|---|---|
| `SpecValid` | Spec has been validated |
| `DeploymentFound` | Referenced ArangoDeployment exists |
| `RouteFound` | Referenced ArangoRoute exists and is ready |
| `Ready` | Connector is ready and visible in `/_inventory` |

## Discovery

Once Ready, the connector appears in the `/_inventory` response:

```json
{
  "connectors": {
    "aql-connector": {
      "description": "Execute AQL queries on ArangoDB",
      "tags": ["database", "aql", "query"],
      "schema": "{\"type\":\"object\",\"properties\":{\"query\":{\"type\":\"string\"},...}}",
      "version": "1.0.0"
    }
  }
}
```

AI tools filter connectors by `tags` and use `schema` to validate their input
before submitting jobs.

## Routing with ArangoRoute

Each connector should have an `ArangoRoute` that redirects from a user-friendly
path to the connector's integration endpoint. The route is referenced in the
connector spec via the `route` field, and the handler verifies it exists.

### How Redirection Works

The `ArangoRoute` redirects requests from `/connector/<name>/` to the internal
`/_integration/connector/v1/` endpoint. This means:

```
Client request:     POST /connector/aql-connector/job
  ↓ (ArangoRoute redirect)
Internal endpoint:  POST /_integration/connector/v1/job
```

All connector API paths are relative, so the redirect works transparently:

| Client path | Redirected to |
|---|---|
| `/connector/aql-connector/job` | `/_integration/connector/v1/job` |
| `/connector/aql-connector/job/{id}` | `/_integration/connector/v1/job/{id}` |
| `/connector/aql-connector/job/{id}/cancel` | `/_integration/connector/v1/job/{id}/cancel` |

### Full Example

Create the route alongside the connector:

```yaml
apiVersion: networking.arangodb.com/v1beta1
kind: ArangoRoute
metadata:
  name: aql-connector-route
spec:
  deployment: my-deployment
  route:
    path: /connector/aql-connector/
  destination:
    path: /_integration/connector/v1/
---
apiVersion: platform.arangodb.com/v1beta1
kind: ArangoPlatformConnector
metadata:
  name: aql-connector
spec:
  type: Active
  deployment:
    name: my-deployment
  route:
    name: aql-connector-route
  description: "Execute AQL queries on ArangoDB"
  tags:
    - database
    - aql
  schema:
    type: object
    properties:
      query:
        type: string
    required:
      - query
  version: "1.0.0"
```

The handler checks the `RouteFound` condition — the connector only becomes
fully `Ready` when the referenced route exists and is active.

The Helm chart should create both the `ArangoPlatformConnector` and its
`ArangoRoute` together.
