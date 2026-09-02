---
layout: page
title: Integration Sidecar Storage V2
grand_parent: ArangoDBPlatform
parent: Integration Sidecars
---

# Storage V2

Definitions:

- [Service](https://github.com/arangodb/kube-arangodb/blob/1.4.5/integrations/storage/v2/definition/storage.proto)

## Configuration

In order to configure Platform Storage, refer to the [documentation](../platform/storage.md).

## RBAC Permissions

Every Storage V2 object operation is authorized against the [Authorization V1](authorization.v1.md)
service before it runs: the caller's token must be granted the matching action on the object path, via
an [`ArangoPermissionPolicy`](../platform/rbac/policies.md) bound to their role. A denied check fails
the request.

| Action | Resource |
|---|---|
| `storage:WriteObject` | the object path being written |
| `storage:ReadObject` | the object path being read |
| `storage:HeadObject` | the object path being stat-ed |
| `storage:DeleteObject` | the object path being deleted |
| `storage:ListObjects` | the path prefix being listed |
| `storage:Init` | *(empty)* — bucket-level initialization |

Example policy statement granting read-only access to objects under `reports/`:

```yaml
statements:
  - effect: Allow
    actions:
      - "storage:ReadObject"
      - "storage:HeadObject"
      - "storage:ListObjects"
    resources:
      - "reports/*"
```
