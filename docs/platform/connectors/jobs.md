---
layout: page
title: Jobs
parent: Connectors
grand_parent: ArangoDBPlatform
nav_order: 3
---

# Connector Jobs

A job represents a unit of work submitted to a connector.

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
| **Scheduled** | Connector claimed the job |
| **Running** | Executing on remote source |
| **Completed** | Done, results in FileStore |
| **Failed** | Failed, description explains why |
| **Cancelled** | Cancelled by AI tool |

## Job Lifecycle

1. **AI tool** creates a job with a query matching the connector's schema
2. Job is stored with status **Pending**
3. **Connector** picks up the job — status moves to **Scheduled**
4. Connector updates to **Running** when execution begins
5. Connector uploads results to FileStore
6. Connector updates to **Completed** (or **Failed**)
7. AI tool polls the job, reads results from the `result` path

## Status History

Each job keeps a status history (up to 10 entries, most recent first):

```json
{
  "status": [
    {"state": "JOB_STATE_COMPLETED", "description": "Query returned 42 docs", "updated_at": "..."},
    {"state": "JOB_STATE_RUNNING", "description": "Executing AQL", "updated_at": "..."},
    {"state": "JOB_STATE_SCHEDULED", "description": "Job scheduled", "updated_at": "..."},
    {"state": "JOB_STATE_PENDING", "description": "Job created", "updated_at": "..."}
  ]
}
```

## Results

Results are stored in FileStore at:

```
/connectors/<connector-id>/<job-id>/
```

The `result` field on the job contains this path. Use the FileStore API
(StorageV2) to read the uploaded files.

## Timeouts

Timeouts are per-job, set by the AI tool when creating the job.
If no timeout is specified, the job runs until completion or failure.
Timeout handling is the connector's responsibility.
