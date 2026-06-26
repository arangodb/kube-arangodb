---
layout: page
title: Jobs
parent: Links
grand_parent: ArangoDBPlatform
nav_order: 3
---

# Link Jobs

A job represents a single unit of work submitted to a link — for example,
"run this AQL query" or "search this vector index".

Jobs can be created by **any authenticated HTTP client** — AI tools, scripts,
or end users — via the external API or through the link's `ArangoRoute`
(e.g. `POST /link/aql-link/job`).

## Job States

```
Pending ──► Scheduled ──► Running ──► Completed
  │            │            │
  │            │            ▼
  │            │          Failed
  │            │
  ▼            ▼
Cancelled ◄── Running
```

| State | Description |
|---|---|
| **Pending** | Job submitted, waiting to be picked up |
| **Scheduled** | Link claimed the job |
| **Running** | Executing on remote source |
| **Completed** | Done, results in FileStore |
| **Failed** | Failed, description explains why |
| **Cancelled** | Cancelled by AI tool |

## Job Lifecycle

1. **AI tool** creates a job with an input matching the link's schema
2. Job is stored with status **Pending**
3. **Connector** picks up the job — status moves to **Scheduled**
4. Connector updates to **Running** when execution begins
5. Connector uploads results to FileStore
6. Connector updates to **Completed** (or **Failed**)
7. AI tool polls the job, reads results from the `result` path

## Status History

Each job keeps a status history of **up to 10 entries per job** (most recent
first). This is the history for a single job — not a global limit. There is
currently no hard limit on the total number of jobs; they are stored in
MetaStore (ArangoDB) and persist until explicitly deleted.

```json
{
  "statuses": [
    {"state": "JOB_STATE_COMPLETED", "description": "Query returned 42 docs", "updated": "..."},
    {"state": "JOB_STATE_RUNNING", "description": "Executing AQL", "updated": "..."},
    {"state": "JOB_STATE_SCHEDULED", "description": "Job scheduled", "updated": "..."},
    {"state": "JOB_STATE_PENDING", "description": "Job created", "updated": "..."}
  ]
}
```

### Accessing Job History

**Via API** — the primary way to access job status and history:

```bash
# Get a specific job with full status history
curl https://<gateway>/link/<name>/job/<job-id>

# List all jobs (optionally filter by state)
curl https://<gateway>/link/<name>/job
curl https://<gateway>/link/<name>/job?state=JOB_STATE_FAILED
```

**Via kubectl** — jobs are stored in MetaStore (ArangoDB), not as Kubernetes
resources, so they are not visible via `kubectl`. Use the API endpoints above.

## Results

Results are stored in StorageV2 (FileStore) at:

```
/links/<link-id>/<job-id>/
```

The `result` field on a completed job contains this path prefix. Use the
**StorageV2** gRPC client to list and read result files directly:

```go
// List all result files for a job
objects, err := pbStorageV2.List(ctx, storageClient, job.GetResult())

// Read a specific file
var buf bytes.Buffer
_, err := pbStorageV2.Receive(ctx, storageClient, obj.GetPath().GetPath(), &buf)
```

Files are named by the link — typically `result.<batch>.jsonl` for
connectors that produce JSONL output in batches. The link's `GetInfo`
response documents the expected file naming pattern and output format.

## Timeouts

Timeouts are per-job, set by the AI tool when creating the job.
If no timeout is specified, the job runs until completion or failure.
Timeout handling is the link's responsibility.
