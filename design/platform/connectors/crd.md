# ArangoPlatformConnector CRD

## Resource

```
apiVersion: platform.arangodb.com/v1beta1
kind: ArangoPlatformConnector
```

Group: `platform.arangodb.com`
Plural: `arangoplatformconnectors`

## Spec

| Field | Type | Description |
|---|---|---|
| `description` | `string` | Human-readable description of what the connector does |
| `tags` | `[]string` | Labels for discovery and filtering (e.g. `["database", "aql"]`) |
| `schema` | `string` | JSON Schema defining the connector's input query parameters |
| `version` | `string` | Connector version |

AI tools use these fields to discover connectors via `/_inventory` and
validate their input before submitting jobs.

## Status

| Field | Type | Description |
|---|---|---|
| `conditions` | `ConditionList` | Standard conditions: `SpecValid`, `Ready` |

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
  schema: |
    {
      "type": "object",
      "properties": {
        "query": {
          "type": "string",
          "description": "AQL query string"
        },
        "bindVars": {
          "type": "object",
          "description": "Bind variables for the query"
        }
      },
      "required": ["query"]
    }
  version: "1.0.0"
```

## Discovery

Once the connector is `Ready`, it appears in `/_inventory`:

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

The operator handler (`pkg/handlers/platform/connector/`) reconciles
`ArangoPlatformConnector` resources:

1. Validates the spec
2. Sets `SpecValid` condition
3. Sets `Ready` condition

The handler is registered in the platform operator alongside storage,
chart, and service handlers.
