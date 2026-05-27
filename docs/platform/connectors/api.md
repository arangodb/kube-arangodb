---
layout: page
title: Connector API
parent: Connectors
grand_parent: ArangoDBPlatform
nav_order: 2
---

# Connector API

The connector exposes two sets of endpoints through the integration sidecar.

## External API (AI Tools)

Used by AI tools to submit and manage jobs. Exposed on the external gateway.

### Create Job

```
POST /_integration/connector/v1/job

{
  "query": "<base64 JSON matching connector schema>",
  "timeout": "30s"
}

Response:
{
  "id": "550e8400-e29b-41d4-a716-446655440000"
}
```

### Get Job

```
GET /_integration/connector/v1/job/{id}

Response:
{
  "job": {
    "id": "550e8400-...",
    "connector_id": "...",
    "statuses": [
      {"state": "JOB_STATE_COMPLETED", "description": "Query returned 42 documents", "updated": "..."},
      {"state": "JOB_STATE_RUNNING", "description": "Executing AQL query", "updated": "..."},
      ...
    ],
    "result": "/connectors/<connector-id>/<job-id>/"
  }
}
```

### List Jobs

```
GET /_integration/connector/v1/job
GET /_integration/connector/v1/job?state=JOB_STATE_PENDING

Response:
{
  "jobs": [...]
}
```

### Cancel Job

```
POST /_integration/connector/v1/job/{id}/cancel

Response:
{
  "job": { ... }
}
```

## Internal API (Connector Process)

Used by the connector binary to pick up and process jobs. Only accessible
from within the pod (internal listener).

### Pick Up Job

```
POST /_internal/connector/v1/job/pickup

Response:
{
  "id": "550e8400-..."
}
```

Returns empty `{}` if no pending jobs. Atomically moves the job from
Pending to Scheduled.

### Get Job

```
GET /_internal/connector/v1/job/{id}
```

### Update Job Status

```
POST /_internal/connector/v1/job/{id}/status

{
  "status": {
    "state": "JOB_STATE_RUNNING",
    "description": "Executing AQL query"
  }
}
```

### Upload File

```
POST /_internal/connector/v1/job/{job_id}/upload/{name}

Body: <raw file bytes>

Response:
{
  "bytes": 1234,
  "checksum": "<sha256>"
}
```
