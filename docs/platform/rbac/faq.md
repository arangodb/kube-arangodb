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

### How do wildcards work in actions?

Actions use colon-separated segments matched left to right:

| Pattern | Matches |
|---|---|
| `*` | Everything |
| `database:*` | `database:read`, `database:write`, `database:drop` |
| `*:read` | `database:read`, `collection:read` |
| `database:read` | Only `database:read` |

### Can a Deny override an Allow?

Yes. Explicit Deny always wins regardless of order. If any policy statement
denies the action/resource, the request is denied even if another statement
allows it.

### What is the action format?

Actions follow `<namespace>:<name>`. Common namespaces:

| Namespace | Description |
|---|---|
| `database` | Database-level operations |
| `collection` | Collection-level operations |
| `document` | Document-level operations |
| `rbac` | RBAC management operations |

## Roles and Scopes

### Where is the scope defined?

The scope is defined per user-role assignment, not on the role. A role groups
named policies; the scope boundary is set on the `ArangoPermissionRoleUserBinding`
(or token) that grants the role to a user. Roles no longer carry their own scope.

### Can I assign the same role to different users with different scopes?

Yes. Use the User Role Binding API:

```bash
# Alice gets editor scoped to reports
curl -X POST https://<gw>/_management/permissions/user/alice/role/editor \
  -d '{"scope": {"statements": [{"effect":"Allow","actions":["*"],"resources":["reports"]}]}}'

# Bob gets editor scoped to analytics
curl -X POST https://<gw>/_management/permissions/user/bob/role/editor \
  -d '{"scope": {"statements": [{"effect":"Allow","actions":["*"],"resources":["analytics"]}]}}'
```

## Tokens

### How do I get a JWT token for my service?

Create an `ArangoPermissionPolicy` and an `ArangoPermissionToken` that
references it:

```yaml
apiVersion: permission.arangodb.com/v1alpha1
kind: ArangoPermissionPolicy
metadata:
  name: my-service-policy
spec:
  deployment:
    name: my-deployment
  policy:
    statements:
      - effect: Allow
        actions: ["collection:read"]
        resources: ["*"]
---
apiVersion: permission.arangodb.com/v1alpha1
kind: ArangoPermissionToken
metadata:
  name: my-service
spec:
  deployment:
    name: my-deployment
  roles: [viewer]
  policyName: my-service-policy
  scope:
    statements:
      - effect: Allow
        actions: ["collection:read"]
        resources: ["*"]
```

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
  -d '{"action": "collection:write", "resource": "my-collection"}'
```

### How do I see what roles my token has?

Use the validate endpoint:

```bash
curl -X POST https://<gateway>/_integration/authn/v1/validate \
  -d '{"token": "<jwt-token>"}'
```

This returns the username, roles list, and token lifetime.

## Troubleshooting

### All requests are denied after enabling RBAC

1. Verify you are in `central-permissive` mode first (not `central`)
2. Check that policies and roles exist in the sidecar
3. Verify the JWT contains the expected roles:
   ```bash
   echo "$TOKEN" | cut -d. -f2 | base64 -d 2>/dev/null | jq .
   ```
4. Check operator logs for policy creation errors

### Permission denied but I think I have access

1. Check your identity: `GET /_integration/authn/v1/identity`
2. List your roles from the token: `POST /_integration/authn/v1/validate`
3. Test the specific permission: `POST /_management/permissions/validate`
4. Check if an explicit Deny is overriding your Allow

### Token does not contain the managed role

The managed role (`managed:operator:<uid>`) is added to the JWT after the
role is created in the sidecar. This takes one additional reconcile cycle.
Wait a few seconds and check the token again — it will be regenerated
automatically.

Also verify that the referenced `ArangoPermissionPolicy` (via `spec.policyName`)
exists and is in a Ready state before the token can create its managed role.
