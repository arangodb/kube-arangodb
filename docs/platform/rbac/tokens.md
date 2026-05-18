---
layout: page
title: Permission Tokens
parent: RBAC
grand_parent: ArangoDBPlatform
nav_order: 3
---

# Permission Tokens

The `ArangoPermissionToken` custom resource creates JWT tokens for accessing
ArangoDB deployments with specific roles and policies.

## Creating a Token

```yaml
apiVersion: permission.arangodb.com/v1alpha1
kind: ArangoPermissionToken
metadata:
  name: my-service-token
spec:
  deployment:
    name: my-deployment
  ttl: 2h
  roles:
    - viewer
    - editor
  policyName: my-policy
  scope:
    statements:
      - effect: Allow
        actions:
          - "collection:read"
          - "collection:write"
        resources:
          - "my-collection"
```

## Spec Fields

| Field | Required | Default | Description |
|---|---|---|---|
| `deployment.name` | Yes | - | Name of the ArangoDeployment |
| `ttl` | No | `1h` | Token lifetime (minimum 15m) |
| `roles` | No | `[]` | List of role names to include in the JWT |
| `policyName` | No | - | Name of the `ArangoPermissionPolicy` CRD to reference |
| `scope` | No | - | Boundary policy that constrains what the referenced policy can grant on this token's managed role |

## Scope

The `scope` field defines the **permission boundary** for the token's managed
role. Even if the referenced policy grants broad access, the scope restricts
what the managed role can actually provide.

For example, if `policyName` references a policy that allows `*:*` on `*`, but
the scope only allows `collection:read` on `my-collection`, then the token's
managed role will only grant read access to `my-collection`.

When `scope` is omitted, the managed role has no additional constraints beyond
the referenced policy.

## What the Operator Creates

When `spec.policyName` is set, the operator automatically:

1. **Creates an ArangoDB user** with a random password
2. **Resolves the ArangoPermissionPolicy** CRD referenced by `spec.policyName`
3. **Creates a role** in the sidecar (`managed:operator:<uid>`) that references
   the external policy and uses `spec.scope` as its boundary
4. **Generates a JWT** signed with the deployment's JWT secret, containing:
   - The created username
   - Roles from `spec.roles` plus the managed role name
   - Expiration based on `spec.ttl`
5. **Stores the JWT** in a Kubernetes Secret (owned by the token resource)

## Retrieving the Token

The JWT is stored in a Secret created by the operator:

```bash
kubectl get secret -l "app.kubernetes.io/managed-by=arangodb-operator" \
  -o jsonpath='{.items[0].data.token}' | base64 -d
```

Or check the token status for the secret name:

```bash
kubectl get arangopermissiontoken my-service-token -o jsonpath='{.status.secret.name}'
```

Then:

```bash
kubectl get secret <secret-name> -o jsonpath='{.data.token}' | base64 -d
```

## Token Lifecycle

- Tokens are **automatically refreshed** at half their TTL
- The JWT is **regenerated** when:
  - The signing secret changes
  - The roles list changes (e.g., managed role is created)
  - The TTL changes
- On **deletion**, the operator cleans up:
  - The ArangoDB user
  - The sidecar role

The referenced `ArangoPermissionPolicy` is **not** deleted when the token is
removed — it is managed independently.

## Example: Read-Only Service Token

First create a policy:

```yaml
apiVersion: permission.arangodb.com/v1alpha1
kind: ArangoPermissionPolicy
metadata:
  name: monitoring-policy
spec:
  deployment:
    name: production
  policy:
    statements:
      - effect: Allow
        actions:
          - "database:read"
          - "collection:read"
        resources:
          - "*"
      - effect: Deny
        actions:
          - "database:write"
          - "collection:write"
        resources:
          - "*"
```

Then create a token referencing it:

```yaml
apiVersion: permission.arangodb.com/v1alpha1
kind: ArangoPermissionToken
metadata:
  name: monitoring-token
spec:
  deployment:
    name: production
  ttl: 4h
  policyName: monitoring-policy
  scope:
    statements:
      - effect: Allow
        actions:
          - "database:read"
          - "collection:read"
        resources:
          - "*"
```

This creates a token scoped to read-only access. The scope acts as a boundary —
even if the referenced policy were to be updated with broader permissions, this
token's role would remain limited to read operations.
