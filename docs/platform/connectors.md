---
layout: page
has_children: true
title: Connectors
parent: ArangoDBPlatform
nav_order: 4
---

# Platform Connectors

Connectors allow AI tools to execute operations on remote sources (databases, APIs, services)
through a unified interface.

A connector registers itself via the `ArangoPlatformConnector` CRD, declares its capabilities
(description, tags, JSON Schema), and processes jobs submitted by AI tools.

## Quick Start

1. Deploy a connector (Helm chart or manual)
2. The connector CRD becomes `Ready` and appears in `/_inventory`
3. AI tools discover the connector, read its schema, and submit jobs
4. The connector picks up jobs, executes them, and uploads results

## Resource Model

Connectors are **not serverless** — they run as regular Kubernetes Deployments
that you manage. The connector pod contains:

- **Your connector container** — your binary that processes jobs
- **Integration sidecar** — injected automatically by the platform, provides
  the job queue (MetaStore) and file storage (StorageV2)

The connector runs continuously and polls for jobs. It does **not** run as a
sidecar of the caller — it is an independent workload with its own resource
limits that you configure in the Deployment spec.

You are responsible for:
- Setting CPU/memory limits on the connector container
- Choosing the number of replicas (multiple instances share the job queue safely)
- Monitoring the connector's health

## Sections

- [CRD Reference](connectors/crd.md) — ArangoPlatformConnector resource
- [API Reference](connectors/api.md) — REST and gRPC endpoints
- [Jobs](connectors/jobs.md) — Job lifecycle and states
- [Building a Connector](connectors/building.md) — How to build your own connector
