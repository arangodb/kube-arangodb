---
layout: page
title: Integration Sidecar Authorization V1
grand_parent: ArangoDBPlatform
parent: Integration Sidecars
---

# Authorization V1

The Authorization V1 integration service provides programmatic permission
evaluation endpoints. It is used by other services to check whether a user
is authorized to perform an action on a resource.

## Service Definition

- [Proto](https://github.com/arangodb/kube-arangodb/blob/1.4.5/integrations/authorization/v1/definition/definition.proto)
- [Request Messages](https://github.com/arangodb/kube-arangodb/blob/1.4.5/integrations/authorization/v1/definition/request.proto)

## Endpoints

| Method | Path | Description |
|---|---|---|
| `POST` | `/_integration/authorization/v1/evaluate` | Evaluate a single permission |
| `POST` | `/_integration/authorization/v1/evaluate-many` | Evaluate multiple permissions |
| `POST` | `/_integration/authorization/v1/evaluate-token` | Evaluate from JWT token |
| `POST` | `/_integration/authorization/v1/evaluate-token-many` | Batch evaluate from JWT token |

## Evaluate

Checks if a user with given roles can perform an action on a resource.

Request:
```json
{
  "user": "alice",
  "roles": ["viewer", "editor"],
  "action": "collection:write",
  "resource": "reports"
}
```

Response:
```json
{
  "message": "Access Granted",
  "effect": "Allow"
}
```

## EvaluateMany

Batch version — checks multiple action/resource pairs for the same user.

Request:
```json
{
  "user": "alice",
  "roles": ["viewer"],
  "items": [
    {"action": "collection:read", "resource": "reports"},
    {"action": "collection:write", "resource": "reports"}
  ]
}
```

## EvaluateToken / EvaluateTokenMany

Same as Evaluate/EvaluateMany but takes a JWT token instead of explicit
user and roles. The user and roles are extracted from the token claims.

## RBAC Permissions

Beyond evaluating permissions, the authorization service also exposes the RBAC management API (roles,
policies and user-role bindings) and the streaming pool endpoints that sidecars use to sync RBAC
state. Each management call is itself authorized: the caller's token must be granted the matching
`rbac:*` action, via an [`ArangoPermissionPolicy`](../platform/rbac/policies.md) bound to their role.

| Action | Resource |
|---|---|
| `rbac:ListRole`, `rbac:GetRole`, `rbac:CreateRole`, `rbac:UpdateRole`, `rbac:DeleteRole` | the role name (empty for List) |
| `rbac:ListPolicy`, `rbac:GetPolicy`, `rbac:CreatePolicy`, `rbac:UpdatePolicy`, `rbac:DeletePolicy` | the policy name (empty for List) |
| `rbac:ListUserRoleBinding`, `rbac:AssignUserRole`, `rbac:RemoveUserRole`, `rbac:ReplaceUserRoleScope` | the target user |
| `rbac:PoolRole`, `rbac:PoolPolicy`, `rbac:PoolUserRoleBinding` | *(empty)* — the streaming pool sidecars use to sync RBAC state |

The `Evaluate` / `EvaluateToken` endpoints above are **not** gated by an `rbac:*` action — they are the
enforcement primitive other services call to authorize their own operations.

## Forcing a refresh on evaluation (read-your-writes)

Each sidecar serves `Evaluate` from a locally cached copy of the RBAC pools that it streams from the
pool service, so immediately after a policy/role/binding change a *different* sidecar (for example
another coordinator) may still answer from a slightly stale cache until its stream catches up.

A caller that just changed RBAC state and must not observe that stale cache can set the gRPC request
metadata header `x-arangodb-authorization-refresh: true` on any of `Evaluate`, `EvaluateMany`,
`EvaluateToken` or `EvaluateTokenMany`. The service then re-pulls its pools (policies, roles and
user/group role bindings) from the store **before** evaluating, giving read-your-writes consistency for
that call. It is opt-in because it trades a little latency for consistency; the data-path callers
(arangod / the gateway) do not set it. The header name is exported as
`authorization/v1/definition.RefreshHeader`. It is equivalent to calling the `Refresh` RPC first, but
scoped to the single evaluation and guaranteed to run on whichever sidecar serves it.

## Configuration

The authorization mode is controlled by the `INTEGRATION_AUTHORIZATION_V1_TYPE`
environment variable:

| Value | Behavior |
|---|---|
| `central` | Full policy enforcement |
| `central-permissive` | Evaluate but allow on error |
| `always` | Always allow |
| `never` | Always deny |

See [RBAC](../platform.rbac.md) for details on enabling and configuring
authorization.
