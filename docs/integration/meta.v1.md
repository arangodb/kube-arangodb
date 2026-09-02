---
layout: page
title: Integration Sidecar Meta V1
grand_parent: ArangoDBPlatform
parent: Integration Sidecars
---

# Meta V1

Definitions:

- [Service](https://github.com/arangodb/kube-arangodb/blob/1.4.5/integrations/meta/v1/definition/definition.proto)

## RBAC Permissions

Every Meta V1 operation is authorized against the [Authorization V1](authorization.v1.md) service
before it runs: the caller's token must be granted the matching action on the key it targets, via an
[`ArangoPermissionPolicy`](../platform/rbac/policies.md) bound to their role. A denied check fails the
request.

| Action | Resource |
|---|---|
| `meta:GetKey` | the key being read |
| `meta:UpdateKey` | the key being written |
| `meta:DeleteKey` | the key being deleted |
| `meta:ListKey` | the key prefix being listed |

Example policy statement granting read access to keys under `config/`:

```yaml
statements:
  - effect: Allow
    actions:
      - "meta:GetKey"
      - "meta:ListKey"
    resources:
      - "config/*"
```
