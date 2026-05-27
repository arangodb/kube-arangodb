# Link Service Implementation

## Structure

```
integrations/link/v1/
├── definition/
│   ├── job.proto                  # Job, JobStatus, JobState
│   ├── external.proto             # LinkV1External (REST + gRPC)
│   └── internal.proto             # LinkV1Internal (HTTP + gRPC)
├── implementation.go              # Struct, New(), Register, Health, Background
├── implementation_external.go     # CreateJob, ListJobs, CancelJob
├── implementation_internal.go     # PickUpJob, GetJob, UpdateJobStatus, UploadFile, BatchUploadFiles
├── store.go                       # MetaStore-backed job store
├── handler.go                     # Handler heartbeat loop
└── consts.go                      # Service name

pkg/integrations/
└── connector_v1.go                # Integration sidecar extension registration
```

## Integration Sidecar Extension

The connector registers as an integration extension in `pkg/integrations/connector_v1.go`.

### Listener Configuration

The integration sidecar runs two listeners:

| Listener | Address | Gateway | Purpose |
|---|---|---|---|
| **Internal** | `127.0.0.1:9092` | `127.0.0.1:9192` | Connector processes (local only) |
| **External** | `0.0.0.0:9093` | `0.0.0.0:9193` | AI tools (exposed via gateway) |

The connector implements `IntegrationEnablement` returning `(internal=true, external=true)`,
so both `LinkV1External` and `LinkV1Internal` gRPC services are registered on both
listeners. Route separation is handled by HTTP gateway annotations:

- External gateway (`/_integration/connector/v1/*`) — AI tool endpoints
- Internal gateway (`/_internal/connector/v1/*`) — Connector process endpoints

### CLI Flags

```
--integration.link.v1                          Enable LinkV1 integration
--integration.link.v1.internal                 Enable on internal listener (default: true)
--integration.link.v1.external                 Enable on external listener (default: true)
--integration.link.v1.connector-id             Link UUID (required)
--integration.link.v1.internal-address          Internal gRPC address for MetaV1/StorageV2 clients
```

### Dependencies

The connector creates gRPC clients to co-located services:
- **MetaV1** — for job storage (via internal gRPC address)
- **StorageV2** — for file uploads (via internal gRPC address)

## Initialization

```go
handler := linkV1.New(metaClient, storageClient, linkID)
```

- `metaClient` — MetaV1 gRPC client for job storage
- `storageClient` — StorageV2 gRPC client for file uploads
- `linkID` — UUID from configuration (identifies the link type)
- `handlerID` — generated internally via `uuid.New()`

## gRPC Registration

Both services registered on the same gRPC server:

```go
func (i *implementation) Register(registrar *grpc.Server) {
    pbLinkV1.RegisterLinkV1InternalServer(registrar, i)
    pbLinkV1.RegisterLinkV1ExternalServer(registrar, i)
}
```

## Background

Runs the handler heartbeat loop (blocks until context cancelled):

```go
func (i *implementation) Background(ctx context.Context) {
    startHeartbeat(ctx, i.meta, i.linkID, i.handlerID)
}
```

Heartbeat writes a timestamp to MetaStore every 30 seconds with TTL=1 minute.

## Store

`jobStore` wraps MetaV1 client with:
- Local mutex for within-instance serialization
- Revision-based optimistic locking for cross-instance safety
- State transition validation
- Status history management (max 10 entries)

Key operations:
- `Create` — store new job (no revision)
- `Get` — fetch job + revision
- `PickUp` — list pending, attempt atomic Pending→Scheduled with rev check
- `UpdateStatus` — validate transition, push status, update with rev check
- `Cancel` — validate cancellable state, push Cancelled status with rev check

## File Uploads

Files uploaded to StorageV2 at path: `/links/<link_id>/<job_id>/<filename>`

- `UploadFile` — unary RPC, full data in request body
- `BatchUploadFiles` — client-streaming, new file starts when `name` changes

## Testing

Test mocks in `pkg/util/tests/integration/`:
- `NewMetaV1Client()` — in-memory MetaV1 with revision support
- `NewStorageV2Client(t)` — filesystem-backed StorageV2 with uploads/downloads/files dirs

Test files:
- `suite_test.go` — shared helpers
- `implementation_external_test.go` — external API tests
- `implementation_internal_test.go` — internal API tests
- `handler_test.go` — heartbeat tests
- `lifecycle_test.go` — end-to-end lifecycle tests
