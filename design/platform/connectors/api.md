# Connector API

Two gRPC services served from a single implementation.

## External API — ConnectorV1External

Used by AI tools. Exposed as REST via gateway.

### Create Job

```
POST /_integration/connector/v1/job

{
  "query": "<base64-encoded JSON matching connector schema>",
  "timeout": "30s"
}

Response: { "id": "<job-uuid>" }
```

### Get Job

```
GET /_integration/connector/v1/job/{id}

Response: { "job": { ... } }
```

### List Jobs

```
GET /_integration/connector/v1/job?state=JOB_STATE_PENDING

Response: { "jobs": [ ... ] }
```

### Cancel Job

```
POST /_integration/connector/v1/job/{id}/cancel

Response: { "job": { ... } }
```

## Internal API — ConnectorV1Internal

Used by connector processes. HTTP + gRPC.

### Pick Up Job

```
POST /_internal/connector/v1/job/pickup

Response: { "id": "<job-uuid>" }   // or {} if no pending jobs
```

Atomically moves one Pending job to Scheduled. Sets `handler_id` and
`result` path. Uses MetaStore revision check for concurrency safety.

### Get Job

```
GET /_internal/connector/v1/job/{id}

Response: { "job": { ... } }
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

Response: { "job": { ... } }
```

### Upload File

```
POST /_internal/connector/v1/job/{job_id}/upload/{name}

Body: <raw file bytes>

Response: { "bytes": 1234, "checksum": "<sha256>" }
```

### Batch Upload Files (gRPC only)

Client-streaming RPC. Each file starts with a message containing `job_id`
and `name`, followed by data chunks. Returns results per file.

## Proto Files

| File | Contents |
|---|---|
| `job.proto` | `Job`, `JobStatus`, `JobState` |
| `external.proto` | `ConnectorV1External` service + request/response messages |
| `internal.proto` | `ConnectorV1Internal` service + request/response messages |
