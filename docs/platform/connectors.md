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

## Sections

- [CRD Reference](connectors/crd.md) — ArangoPlatformConnector resource
- [API Reference](connectors/api.md) — REST and gRPC endpoints
- [Jobs](connectors/jobs.md) — Job lifecycle and states
- [Building a Connector](connectors/building.md) — How to build your own connector
