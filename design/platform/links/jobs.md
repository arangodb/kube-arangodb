# Link Jobs

## Job States

```
Pending ──► Scheduled ──► Running ──► Completed
  │            │            │
  │            │            ▼
  │            │          Failed (with description)
  │            │
  ▼            ▼
Cancelled ◄── Running
```

| State | Description |
|---|---|
| **Pending** | Job submitted, waiting to be picked up |
| **Scheduled** | Handler claimed the job, `handler_id` and `result` path set |
| **Running** | link is executing on remote source |
| **Completed** | Done, results available in FileStore |
| **Failed** | Execution failed (includes timeout), `description` explains why |
| **Cancelled** | Cancelled by AI tool from Pending, Scheduled, or Running |

### Valid Transitions

| From | To |
|---|---|
| Pending | Scheduled (via PickUpJob) |
| Pending | Cancelled (via CancelJob) |
| Scheduled | Running |
| Scheduled | Failed |
| Scheduled | Cancelled |
| Running | Completed |
| Running | Failed |
| Running | Cancelled |

## Job Fields

| Field | Type | Set when |
|---|---|---|
| `id` | UUID | Created |
| `link_id` | UUID | Created (from config) |
| `handler_id` | UUID | Scheduled (handler instance) |
| `statuses` | `[]JobStatus` | Every transition (max 10, newest first) |
| `query` | bytes (JSON) | Created |
| `timeout` | Duration | Created (optional) |
| `created` | Timestamp | Created |
| `result` | string | Scheduled (`/links/<cid>/<jid>/`) |

## Status History

Status is a list (max 10 entries, most recent first). Each entry:

| Field | Type |
|---|---|
| `state` | JobState |
| `description` | string |
| `updated` | Timestamp |

Example after completion:
```json
{
  "statuses": [
    { "state": "JOB_STATE_COMPLETED", "description": "Query returned 42 documents", "updated": "..." },
    { "state": "JOB_STATE_RUNNING", "description": "Executing AQL query", "updated": "..." },
    { "state": "JOB_STATE_SCHEDULED", "description": "Job scheduled", "updated": "..." },
    { "state": "JOB_STATE_PENDING", "description": "Job created", "updated": "..." }
  ]
}
```

## Storage

### MetaStore

Jobs stored via MetaV1 gRPC client (ArangoDB-backed).

Key: `links/<link_id>/jobs/<job_id>`

Concurrency: revision-based optimistic locking. PickUp lists pending jobs
and attempts Pending→Scheduled with rev check. On revision conflict (another
handler claimed it), skips to next job.

Local mutex serializes operations within a single handler instance.

### FileStore

Results stored at: `/links/<link_id>/<job_id>/`

Path assigned when job moves to Scheduled. Connector uploads via
`UploadFile` (unary, full data) or `BatchUploadFiles` (streaming).
Backed by StorageV2 service (S3/GCS/Azure).

## Handler Heartbeat

Each handler instance registers at:
`links/<link_id>/handlers/<handler_id>`

- TTL: 1 minute
- Renewed: every 30 seconds
- On crash: entry expires, dead handler detectable

### Identity

| ID | Scope | Generated |
|---|---|---|
| ConnectorUUID | per link type | from configuration |
| HandlerUUID | per runtime instance | on startup (`uuid.New()`) |
