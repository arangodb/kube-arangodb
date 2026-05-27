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

| Field | Type | Description |
|---|---|---|
| `description` | `string` | Human-readable description of what the connector does |
| `tags` | `[]string` | Labels for discovery and filtering |
| `schema` | `JSONSchemaProps` | [JSON Schema](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/#specifying-a-structural-schema) defining the connector's query input parameters |
| `version` | `string` | Connector version |

## Status

| Condition | Description |
|---|---|
| `SpecValid` | Spec has been validated |
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
