# Platform Links

## Overview

Links are an abstraction layer that allows AI tools to execute operations
on remote sources (databases, APIs, services) through a unified interface.

A link is a plugin that registers itself with the platform via the
`ArangoPlatformLink` CRD, declares its capabilities, and processes
jobs submitted by AI tools.

## Architecture

```
  AI Tool                      Platform                         Remote Source
  ------                       --------                         -------------

  1. Discover        ──►  /_inventory
                          (connectors with tags, schema)

  2. Submit job      ──►  LinkV1External
                          POST /_integration/connector/v1/job
                                    │
                                    ▼
                              MetaStore (Pending)
                                    │
  3. Connector polls ──►  LinkV1Internal
                          POST /_internal/connector/v1/job/pickup
                                    │
                                    ▼
                              MetaStore (Scheduled)
                                    │
  4. Execute         ──────────────────────────────────────►  Query/API call
                                    │
  5. Upload results  ──►  POST /_internal/connector/v1/job/{id}/upload/{name}
                                    │
                                    ▼
                              FileStore (/links/<cid>/<jid>/)
                                    │
  6. Complete        ──►  POST /_internal/connector/v1/job/{id}/status
                                    │
                                    ▼
                              MetaStore (Completed)

  7. Poll + fetch    ──►  GET /_integration/connector/v1/job/{id}
                          GET FileStore results
```

## Listener Architecture

The integration sidecar runs two listeners with separate gateways:

```
External Listener (0.0.0.0:9093)          Internal Listener (127.0.0.1:9092)
├── Gateway (0.0.0.0:9193)                ├── Gateway (127.0.0.1:9192)
│   └── /_integration/connector/v1/*      │   └── /_internal/connector/v1/*
│       ├── POST /job          (Create)   │       ├── POST /job/pickup    (PickUp)
│       ├── GET  /job          (List)     │       ├── GET  /job/{id}      (GetJob)
│       ├── GET  /job/{id}     (Get)      │       ├── POST /job/{id}/status (Update)
│       └── POST /job/{id}/cancel         │       └── POST /job/{id}/upload/{name}
│                                         │
└── LinkV1External (gRPC)            └── LinkV1Internal (gRPC)
                                              └── BatchUploadFiles (streaming)
```

Both gRPC services are registered on both listeners. HTTP route separation is
handled by gateway annotations (`/_integration/*` vs `/_internal/*`).

## Components

| Component | Location | Purpose |
|---|---|---|
| [CRD](crd.md) | `pkg/apis/platform/v1beta1/connector*.go` | ArangoPlatformLink definition |
| [API](api.md) | `integrations/link/v1/definition/` | Proto definitions (External + Internal) |
| [Service](service.md) | `integrations/link/v1/` | gRPC service implementation + sidecar extension |
| [Jobs](jobs.md) | — | Job lifecycle, states, storage |
| Sample | `modules/test/tests/links/aql/` | Sample AQL connector with Helm chart |
