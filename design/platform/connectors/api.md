# Connector API

The connector system provides two gRPC services served from a single
implementation. Both are registered on the integration sidecar's gRPC server.

## ConnectorV1External — AI Tool / End User API

**Purpose**: Allows AI tools, scripts, or end users to submit work and track
results. This is the "consumer" API.

**Audience**: Any authenticated HTTP client that can reach the deployment's
gateway. When an `ArangoRoute` is configured, users access it at
`/connector/<name>/`. Without a route, use `/_integration/connector/v1/`
directly on the gateway.

**All operations are asynchronous**: you create a job, then poll for completion.
There is no push/notification mechanism currently — the client must poll
`GET /job/{id}` to check progress. A push-on-change (e.g. WebSocket or
server-sent events) mechanism may be considered in the future if polling
proves insufficient.

### Create Job

Submit a new job. The `query` field must match the connector's JSON Schema.

```
POST /_integration/connector/v1/job

{
  "query": "<JSON string matching connector schema>",
  "timeout": "30s"
}

Response: { "id": "<job-uuid>" }
```

**Note on encoding**: The `query` field is a proto `bytes` field. In JSON
serialization, proto bytes are base64-encoded. If you're calling via the
HTTP gateway, send the query as a JSON string — the gateway handles encoding.
If you're calling via gRPC directly, send raw bytes.

The job starts in `Pending` state. The connector will pick it up asynchronously.

### Get Job

Retrieve job state, status history, and result path.

```
GET /_integration/connector/v1/job/{id}

Response:
{
  "job": {
    "id": "...",
    "statuses": [{"state": "JOB_STATE_COMPLETED", ...}],
    "result": "/connectors/<cid>/<jid>/"
  }
}
```

### List Jobs

List all jobs, optionally filtered by state.

```
GET /_integration/connector/v1/job?state=JOB_STATE_PENDING

Response: { "jobs": [ ... ] }
```

### Cancel Job

Cancel a Pending, Scheduled, or Running job. Cannot cancel Completed or Failed.

```
POST /_integration/connector/v1/job/{id}/cancel

Response: { "job": { ... } }
```

## ConnectorV1Internal — Connector Process API

**Purpose**: Allows the connector binary to claim jobs from the queue, report
progress, and upload results. This is the "producer" API.

**Audience**: Only the connector process running inside the same pod. Accessible
via the internal listener at `127.0.0.1:9192` — not reachable from outside the pod.

### Pick Up Job

Atomically claim one pending job. Moves it to Scheduled state, assigns the
handler instance, and sets the result FileStore path.

```
POST /_internal/connector/v1/job/pickup

Response: { "id": "<job-uuid>" }   // or {} if no pending jobs
```

Uses MetaStore revision check — safe with multiple connector replicas.

### Get Job

Get full job details including the `query` payload to execute.

```
GET /_internal/connector/v1/job/{id}

Response: { "job": { ... } }
```

### Update Job Status

Report progress. Valid transitions: Scheduled→Running, Running→Completed,
Running→Failed, Scheduled→Failed.

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

Upload a result file before marking the job as Completed.

```
POST /_internal/connector/v1/job/{job_id}/upload/{name}

Body: <raw file bytes>

Response: { "bytes": 1234, "checksum": "<sha256>" }
```

### Batch Upload Files (gRPC only)

Client-streaming RPC for uploading multiple files in one call. Not available
via HTTP — gRPC clients only.

## Proto Files

| File | Contents |
|---|---|
| `job.proto` | `Job`, `JobStatus`, `JobState` |
| `external.proto` | `ConnectorV1External` service + request/response messages |
| `internal.proto` | `ConnectorV1Internal` service + request/response messages |
