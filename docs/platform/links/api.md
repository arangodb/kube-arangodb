---
layout: page
title: Link API
parent: Links
grand_parent: ArangoDBPlatform
nav_order: 2
---

# Link API

The connector system provides two separate APIs served by the integration sidecar.
All operations are **asynchronous** — you submit a job, then poll for its completion.

For an explanation of what a job is and how it progresses through states, see
[Jobs](jobs.md).

## External API — for AI Tools and End Users

The External API is used by AI tools (or any HTTP client) to create and manage
jobs. It is exposed on the external gateway and accessible through the
`ArangoRoute` configured for the link.

Each link instance handles its own jobs — when you submit a job, it goes
to the link whose endpoint you called. There is no routing between
connectors.

If you configured an `ArangoRoute` at `/connector/<name>/`, use that path.
Otherwise use the internal integration path directly.

| Route path | Internal path |
|---|---|
| `POST /connector/<name>/job` | `POST /_integration/connector/v1/job` |
| `GET /connector/<name>/job/{id}` | `GET /_integration/connector/v1/job/{id}` |
| `GET /connector/<name>/job` | `GET /_integration/connector/v1/job` |
| `POST /connector/<name>/job/{id}/cancel` | `POST /_integration/connector/v1/job/{id}/cancel` |

### Create Job

Submit work to the link. The `query` field is a JSON object whose
structure must match the link's `schema` (published in `/_inventory`).
The platform validates the query against the schema before accepting the job.

```
POST /_integration/connector/v1/job

{
  "query": "{\"query\": \"FOR d IN users RETURN d\"}",
  "timeout": "30s"
}
```

- `query` — JSON string matching the link's schema. Sent as a string
  (the proto field is `bytes`, which serializes as base64 over JSON).
- `timeout` — optional, maximum time the link has to complete the job.

Response:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000"
}
```

The job starts in `Pending` state. Poll `GET /job/{id}` to track progress.

### Get Job

Retrieve the current state of a job, including its full status history.

```
GET /_integration/connector/v1/job/{id}
```

Response:
```json
{
  "job": {
    "id": "550e8400-...",
    "link_id": "...",
    "statuses": [
      {"state": "JOB_STATE_COMPLETED", "description": "Query returned 42 documents", "updated": "..."},
      {"state": "JOB_STATE_RUNNING", "description": "Executing AQL query", "updated": "..."},
      {"state": "JOB_STATE_SCHEDULED", "description": "Job scheduled", "updated": "..."},
      {"state": "JOB_STATE_PENDING", "description": "Job created", "updated": "..."}
    ],
    "result": "/links/<connector-id>/<job-id>/"
  }
}
```

The first entry in `statuses` is the current state. When state is
`JOB_STATE_COMPLETED`, the `result` field contains the FileStore path
where the link uploaded its output files.

### List Jobs

List all jobs, optionally filtering by state.

```
GET /_integration/connector/v1/job
GET /_integration/connector/v1/job?state=JOB_STATE_PENDING
```

Response:
```json
{
  "jobs": [...]
}
```

### Cancel Job

Cancel a job that is in Pending, Scheduled, or Running state.
Jobs that are already Completed or Failed cannot be cancelled.

```
POST /_integration/connector/v1/job/{id}/cancel
```

Response:
```json
{
  "job": { ... }
}
```

## Internal API — for the link Process

The Internal API is used by the link binary running inside the pod.
It is only accessible from within the pod via the internal listener
(`127.0.0.1:9192`). External users cannot reach these endpoints.

### Pick Up Job

Claim one pending job. This atomically moves the job from Pending to Scheduled
and assigns the current handler instance. Returns empty `{}` if no jobs are
waiting.

```
POST /_internal/connector/v1/job/pickup

Response: { "id": "550e8400-..." }
```

### Get Job

Same as the external Get Job — returns full job details including the query
payload that the link needs to execute.

```
GET /_internal/connector/v1/job/{id}
```

### Update Job Status

Report progress. The connector calls this to move the job through states:
Scheduled → Running → Completed (or Failed).

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

Upload a result file to the job's FileStore directory. Call this before
marking the job as Completed.

```
POST /_internal/connector/v1/job/{job_id}/upload/{name}

Body: <raw file bytes>

Response:
{
  "bytes": 1234,
  "checksum": "<sha256>"
}
```
