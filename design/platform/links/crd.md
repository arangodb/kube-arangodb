# ArangoPlatformLink CRD

## Resource

```
apiVersion: platform.arangodb.com/v1beta1
kind: ArangoPlatformLink
```

Group: `platform.arangodb.com`
Plural: `arangoplatformlinks`

## Spec

| Field | Type | Description |
|---|---|---|
| `description` | `string` | Human-readable description of what the link does |
| `tags` | `[]string` | Labels for discovery and filtering (e.g. `["database", "aql"]`) |
| `schema` | `string` | JSON Schema defining the link's input query parameters |
| `version` | `string` | Connector version |

AI tools use these fields to discover connectors via `/_inventory` and
validate their input before submitting jobs.

### Spec

| Field | Type | Default | Description |
|---|---|---|---|
| `type` | `string` | `Active` | Link pattern type |
| `deployment` | `Object` | | Reference to ArangoDeployment |
| `route` | `Object` | | Reference to ArangoRoute for redirection |
| `description` | `string` | | Human-readable description |
| `tags` | `[]string` | | Discovery/filtering labels |
| `schema` | `JSONSchemaProps` | | JSON Schema for input query |
| `version` | `string` | | Connector version |

### Status

| Condition | Description |
|---|---|
| `SpecValid` | Spec validated (type is supported) |
| `DeploymentFound` | Referenced ArangoDeployment exists |
| `RouteFound` | Referenced ArangoRoute exists and is ready |
| `Ready` | All conditions met |

### Routing

The `route` field references an ArangoRoute that redirects `/connector/<name>/`
to `/_integration/connector/v1/`. This gives each connector a clean URL.

## Example

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
kind: ArangoPlatformLink
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
      bindVars:
        type: object
    required:
      - query
  version: "1.0.0"
```

## Discovery

Once the link is `Ready`, it appears in `/_inventory`:

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

## Handler

The operator handler (`pkg/handlers/platform/link/`) reconciles
`ArangoPlatformLink` resources:

1. Validates the spec
2. Sets `SpecValid` condition
3. Sets `Ready` condition

The handler is registered in the platform operator alongside storage,
chart, and service handlers.
