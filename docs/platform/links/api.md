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

If you configured an `ArangoRoute` at `/link/<name>/`, use that path.
Otherwise use the internal integration path directly.

| Route path | Internal path |
|---|---|
| `POST /link/<name>/job` | `POST /_integration/connector/v1/job` |
| `GET /link/<name>/job/{id}` | `GET /_integration/connector/v1/job/{id}` |
| `GET /link/<name>/job` | `GET /_integration/connector/v1/job` |
| `POST /link/<name>/job/{id}/cancel` | `POST /_integration/connector/v1/job/{id}/cancel` |
| `GET /link/<name>/info` | `GET /_integration/connector/v1/info` |

### Create Job

Submit work to the link. The `input` field is a JSON object whose
structure must match the link's `input_schema` (published in `/_inventory`).
The platform validates the input against the schema before accepting the job.

```
POST /_integration/connector/v1/job

{
  "input": "{\"query\": \"FOR d IN users RETURN d\"}",
  "timeout": "30s"
}
```

- `input` — JSON string matching the link's input schema. Sent as a string
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
  "id": "550e8400-...",
  "link_id": "...",
  "statuses": [
    {"state": "JOB_STATE_COMPLETED", "description": "Query returned 42 documents", "updated": "..."},
    {"state": "JOB_STATE_RUNNING", "description": "Executing AQL query", "updated": "..."},
    {"state": "JOB_STATE_SCHEDULED", "description": "Job scheduled", "updated": "..."},
    {"state": "JOB_STATE_PENDING", "description": "Job created", "updated": "..."}
  ],
  "result": "/links/<link-id>/<job-id>/"
}
```

The first entry in `statuses` is the current state. When state is
`JOB_STATE_COMPLETED`, the `result` field contains the StorageV2 path
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

### Get Info

Retrieve the link's self-describing tool definition. AI agents call this
to learn the link's input schema, output format, and see usage examples
before submitting jobs.

```
GET /_integration/connector/v1/info
```

Response:
```json
{
  "info": {
    "description": "Execute AQL queries on ArangoDB",
    "tags": ["database", "aql", "query"],
    "input_schema": "{ JSON Schema for query format }",
    "output_schema": "{ JSON Schema for result format }",
    "examples": [
      {"name": "Simple query", "input": "{\"query\": \"RETURN 1\"}", "output": "1\n"}
    ],
    "result_files": ["result.0000000.jsonl"]
  }
}
```

### Accessing Result Files

Result files are stored in StorageV2 under the path returned in the job's
`result` field. Use the **StorageV2** gRPC client to list and read them:

```go
// List files under the job's result prefix
objects, err := pbStorageV2.List(ctx, storageClient, job.GetResult())

// Read a specific file
var buf bytes.Buffer
_, err := pbStorageV2.Receive(ctx, storageClient, objectPath, &buf)
```

See [Jobs — Results](jobs.md#results) for details on file naming and layout.

## Internal API — for the link Process

The Internal API is used by the link binary running inside the pod.
It is accessible via gRPC at `INTEGRATION_SERVICE_ADDRESS` (default
`127.0.0.1:9201`) or via HTTP gateway at `INTEGRATION_HTTP_ADDRESS_FULL`
(default `http://127.0.0.1:9203`). External users cannot reach these
endpoints.

**Recommended**: Use the gRPC client (`LinkV1InternalClient`) for type
safety and streaming support. The HTTP gateway is available as a fallback.

### Pick Up Job

Claim one pending job. This atomically moves the job from Pending to Scheduled
and assigns the current handler instance. Returns empty `{}` if no jobs are
waiting.

```
POST /_internal/connector/v1/job/pickup

Response: { "id": "550e8400-..." }
```

### Get Job

Same as the external Get Job — returns full job details including the input
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

### Batch Upload Files (gRPC only)

Upload multiple files in a single streaming RPC. Each file starts with a
message containing `job_id` and `name`; subsequent messages for the same
file contain only `chunk` data. Start a new file by sending a message with
a different `name`.

```protobuf
rpc BatchUploadFiles(stream BatchUploadFileRequest) returns (BatchUploadFilesResponse);
```

This is the recommended way to upload results when processing data in
batches — wrap the stream in an `io.Writer` per file and write directly
from your processing pipeline.

### Update Info

Register the link's tool definition with the sidecar. The request body
is `LinkInfo` directly (not wrapped). This must be called at startup
before the link can accept jobs.

```
POST /_internal/connector/v1/info

{
  "description": "Execute AQL queries on ArangoDB",
  "tags": ["database", "aql", "query"],
  "input_schema": "{ JSON Schema }",
  "output_schema": "{ JSON Schema }",
  "examples": [
    {"name": "Simple query", "input": "{\"query\": \"RETURN 1\"}", "output": "1\n"}
  ],
  "result_files": ["result.0000000.jsonl"]
}
```

Fields:

| Field | Type | Description |
|-------|------|-------------|
| `description` | string | Human-readable description of what the link does |
| `tags` | string[] | Tags for discovery and filtering |
| `input_schema` | string | JSON Schema for `CreateJobRequest.input` |
| `output_schema` | string | JSON Schema for result file format |
| `examples` | LinkExample[] | Sample input/output pairs for agent few-shot learning |
| `result_files` | string[] | File names produced by a typical job |
