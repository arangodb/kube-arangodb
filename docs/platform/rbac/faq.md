---
layout: page
title: FAQ
parent: RBAC
grand_parent: ArangoDBPlatform
nav_order: 7
---

# RBAC FAQ

## General

### How do I check if RBAC is enabled?

**From the UI/API** — query `/_inventory` and check the `security` field:

```bash
curl https://<gateway>/_inventory
```

```json
{
  "security": {
    "authentication": "AuthenticationNative",
    "authorization": "AuthorizationRBAC"
  }
}
```

| `authorization` value | Meaning |
|---|---|
| `AuthorizationNone` | No authorization (auth disabled) |
| `AuthorizationNative` | ArangoDB built-in user permissions |
| `AuthorizationRBAC` | Platform RBAC is active |

The UI should show/hide the RBAC section based on
`security.authorization == "AuthorizationRBAC"`.

**From Kubernetes** — check the `GatewaySidecarEnabled` condition:

```bash
kubectl get arangodeployment <name> -o jsonpath='{.status.conditions}' | \
  jq '.[] | select(.type == "GatewaySidecarEnabled")'
```

### What happens if RBAC is enabled but no policies exist?

All requests are **denied** (deny-by-default). You must create at least one
policy and role to grant access. Start with `central-permissive` mode to
avoid lockouts.

### Can I use RBAC without authentication?

No. RBAC requires authentication to be enabled on the ArangoDeployment
(`spec.authentication.jwtSecretName` must not be `"None"`). Without
authentication, there is no identity to evaluate policies against.

### What is the difference between `central` and `central-permissive`?

- **`central`** — Full enforcement. Denied requests return HTTP 403.
- **`central-permissive`** — Policies are evaluated and denials are logged,
  but requests are still allowed. Use this for testing policies before
  enforcing them.

## Policies

### How do wildcards work?

Actions and resources are split into colon-separated segments and matched
segment by segment. A segment can be `*`, `<prefix>*`, `*<suffix>`, or an
exact value. A trailing `*` segment also matches any further segments.

| Pattern | Matches |
|---|---|
| `*` | Everything |
| `db:*` | Every ArangoDB action, e.g. `db:Read`, `db:WriteData`, `db:Drop` |
| `db:Admin*` | `db:AdminBackup`, `db:AdminReadLogs`, ... |
| `*:Read` | `db:Read` |
| `db:Read` | Only `db:Read` |
| `db:collection:sales:*` | Every collection in database `sales` |

See [Pattern Matching](policies.md#pattern-matching) for details.

### Can a Deny override an Allow?

Yes. Explicit Deny always wins regardless of order. If any policy statement
denies the action/resource, the request is denied even if another statement
allows it.

### What is the action format?

Actions follow `<namespace>:<name>`. Each component checks actions in its own
namespace:

| Namespace | Description |
|---|---|
| `db` | ArangoDB operations, e.g. `db:Read`, `db:WriteData`, `db:Create` |
| `rbac` | RBAC management operations, e.g. `rbac:ListRole` |
| `ai` | AI services and Arango Control Plane, e.g. `ai:acp.secret.read` |

ArangoDB resources are paths like `db:database:<db>` or
`db:collection:<db>:<collection>`. See [Policies and Roles](policies.md#actions)
for the full list.

### Why are all my ArangoDB requests denied, even though the role allows them?

ArangoDB checks `db:UseApiVersion` on `db:apiversion:<version>` for every
request. If the binding scope only lists databases and collections, this check
fails. Add `db:apiversion:*` to the scope's resources.

## Roles and Scopes

### Where is the scope defined?

The scope is defined per user-role assignment, not on the role. A role groups
named policies; the scope boundary is set on the `ArangoPermissionRoleUserBinding`
(or token) that grants the role to a user. Roles no longer carry their own scope.

### Can I assign the same role to different users with different scopes?

Yes. Use the User Role Binding API:

```bash
# Alice gets editor scoped to reports
curl -X POST -H "Authorization: Bearer $TOKEN" https://<gw>/_management/permissions/user/alice/role/editor \
  -d '{"scope": {"statements": [{"effect":"Allow","actions":["*"],"resources":["db:apiversion:*","db:database:reports","db:collection:reports:*"]}]}}'

# Bob gets editor scoped to analytics
curl -X POST -H "Authorization: Bearer $TOKEN" https://<gw>/_management/permissions/user/bob/role/editor \
  -d '{"scope": {"statements": [{"effect":"Allow","actions":["*"],"resources":["db:apiversion:*","db:database:analytics","db:collection:analytics:*"]}]}}'
```

## Tokens

### How do I get a JWT token for my service?

Create an `ArangoPermissionToken` with an inline policy and a scope, and
optionally bind existing roles:

```yaml
apiVersion: permission.arangodb.com/v1alpha1
kind: ArangoPermissionToken
metadata:
  name: my-service
spec:
  deployment:
    name: my-deployment
  policy:
    statements:
      - effect: Allow
        actions: ["db:UseApiVersion", "db:Read"]
        resources: ["*"]
  scope:
    statements:
      - effect: Allow
        actions: ["db:UseApiVersion", "db:Read"]
        resources: ["*"]
  roles:
    - role:
        name: viewer            # ArangoPermissionRole CRD
      scope:
        policy:
          statements:
            - effect: Allow
              actions: ["*"]
              resources: ["*"]
```

Without a `scope`, the token's inline policy grants nothing.

The token is stored in a Kubernetes Secret. See
[Permission Tokens](tokens.md) for details.

### What is the minimum token TTL?

15 minutes. The default is 1 hour. Tokens are automatically refreshed at
half their TTL.

### How do I check my current permissions?

Use the self-check endpoint:

```bash
curl -X POST -H "Authorization: Bearer $TOKEN" \
  https://<gateway>/_management/permissions/validate \
  -d '{"action": "db:WriteData", "resource": "db:collection:sales:orders"}'
```

### How do I see what roles I have?

Roles are not stored in the JWT. They come from the role bindings of your
user and of the groups in your token's `groups` claim. List the bindings
(requires `rbac:ListUserRoleBinding` / `rbac:ListGroupRoleBinding`):

```bash
curl -H "Authorization: Bearer $TOKEN" \
  https://<gateway>/_management/permissions/user/<username>/role
curl -H "Authorization: Bearer $TOKEN" \
  https://<gateway>/_management/permissions/group/<group>/role
```

To see the username, groups, and lifetime of a token, use the validate
endpoint:

```bash
curl -X POST -H "Authorization: Bearer $TOKEN" \
  https://<gateway>/_integration/authn/v1/validate \
  -d '{"token": "<jwt-token>"}'
```

The `roles` field in its response contains the token's `groups` claim.

## Troubleshooting

### All requests are denied after enabling RBAC

1. Verify you are in `central-permissive` mode first (not `central`)
2. Check that policies and roles exist in the sidecar
3. Verify the JWT carries the expected user and groups:
   ```bash
   echo "$TOKEN" | cut -d. -f2 | base64 -d 2>/dev/null | jq .
   ```
4. Check that every role binding has a scope that allows the action. A
   binding with an empty scope grants nothing.
5. Check operator logs for policy creation errors

### Permission denied but I think I have access

1. Check your identity: `GET /_identity`
2. List your role bindings: `GET /_management/permissions/user/<username>/role`
3. Test the specific permission: `POST /_management/permissions/validate`
4. Check if an explicit Deny is overriding your Allow
5. If you just changed a binding, wait for propagation (up to ~30 seconds,
   see [Runtime Behavior](behavior.md)) or call
   `POST /_management/permissions/refresh`

### My permission token grants nothing

Roles are not part of the JWT. The operator creates user role bindings for the
token's user. Check:

- `kubectl get arangopermissiontoken <name> -o yaml`: the token is `Ready`
  and `status.role` / `status.managedPolicy` are set if you use `spec.policy`
- `spec.scope` is set when using `spec.policy`. Without it, the managed role
  is bound with an empty scope and grants nothing.
- every `spec.roles[]` entry has a scope that allows the required actions
- the bindings exist: `GET /_management/permissions/user/<token-user>/role`
