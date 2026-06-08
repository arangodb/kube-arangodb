---
layout: page
has_children: true
title: Links
parent: ArangoDBPlatform
nav_order: 4
---

# Platform Links

Links allow AI tools to execute operations on remote sources (databases, APIs, services)
through a unified interface.

A connector registers itself via the `ArangoPlatformLink` CRD, declares its capabilities
(description, tags, JSON Schema), and processes jobs submitted by AI tools.

## Quick Start

1. Deploy a link (Helm chart or manual)
2. The Link CRD becomes `Ready` and appears in `/_inventory`
3. AI tools discover the link, read its schema, and submit jobs
4. The connector picks up jobs, executes them, and uploads results

## Resource Model

Links are **not serverless** — they run as regular Kubernetes Deployments
that you manage. The link pod contains:

- **Your link container** — your binary that processes jobs
- **Integration sidecar** — injected automatically by the platform, provides
  the job queue (MetaStore) and file storage (StorageV2)

The link runs continuously and polls for jobs. It does **not** run as a
sidecar of the caller — it is an independent workload with its own resource
limits that you configure in the Deployment spec.

You are responsible for:
- Setting CPU/memory limits on the link container
- Choosing the number of replicas (multiple instances share the job queue safely)
- Monitoring the link's health

## Sections

- [CRD Reference](links/crd.md) — ArangoPlatformLink resource
- [API Reference](links/api.md) — REST and gRPC endpoints
- [Jobs](links/jobs.md) — Job lifecycle and states
- [Building a link](links/building.md) — How to build your own connector
