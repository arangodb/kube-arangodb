---
layout: page
title: Permission Tokens
parent: RBAC
grand_parent: ArangoDBPlatform
nav_order: 4
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
    - role:
        name: editor            # ArangoPermissionRole CRD
      scope:
        policy:
          statements:
            - effect: Allow
              actions: ["*"]
              resources: ["db:apiversion:*", "db:database:sales", "db:collection:sales:*"]
  policy:
    statements:
      - effect: Allow
        actions:
          - "db:UseApiVersion"
        resources:
          - "*"
      - effect: Allow
        actions:
          - "db:Read"
          - "db:WriteData"
        resources:
          - "db:database:sales"
          - "db:collection:sales:orders"
  scope:
    statements:
      - effect: Allow
        actions:
          - "*"
        resources:
          - "db:apiversion:*"
          - "db:database:sales"
          - "db:collection:sales:orders"
```

## Spec Fields

| Field | Required | Default | Description |
|---|---|---|---|
| `deployment.name` | Yes | - | Name of the ArangoDeployment |
| `ttl` | No | `1h` | Token lifetime (minimum `15m`) |
| `roles` | No | `[]` | Existing roles to bind to the token's user, each with its own scope (see below) |
| `policy` | No | - | Inline policy. The operator creates it as a managed policy and wraps it in a managed role for this token |
| `scope` | No | - | Boundary for the managed role created from `policy` |

### `roles`

Each entry binds an existing role to the token's user:

| Field | Required | Description |
|---|---|---|
| `role.name` | One of `name`/`direct` | Name of an `ArangoPermissionRole` CRD |
| `role.direct` | One of `name`/`direct` | Exact sidecar role name, e.g. a predefined role `managed:predefined:coredb-reader` |
| `scope.policy` | One of `policy`/`ref` | Inline boundary policy for this binding |
| `scope.ref.name` | One of `policy`/`ref` | Name of an `ArangoPermissionPolicy` CRD to use as the boundary |

Both `role` and `scope` are required for every entry.

## Scope

The `scope` field defines the **permission boundary** for the token's managed
role (the one created from `spec.policy`). An action is granted only if both
the policy **and** the scope allow it. For example, if `policy` allows `*` on
`*` but the scope only allows `db:UseApiVersion` and `db:Read` on
`db:apiversion:*` and `db:collection:sales:orders`, the managed role only
grants read access to the `orders` collection.

> **Note:** RBAC is deny-by-default and a binding's scope must allow an action
> for the binding to grant it. If `spec.policy` is set but `spec.scope` is
> omitted, the managed role is bound with an empty scope and **grants nothing**.
> To give the policy full effect, set a scope that allows everything the policy
> should grant, e.g. `actions: ["*"]`, `resources: ["*"]`.

## What the Operator Creates

1. **An ArangoDB user** with a random password.
2. If `spec.policy` is set:
   - a **managed policy** `managed:operator:<token-uid>:policy` in the sidecar
   - a **managed role** `managed:operator:<token-uid>` referencing that policy
3. **User role bindings** in the sidecar for the created user:
   - every entry of `spec.roles`, with that entry's scope
   - the managed role (if any), with `spec.scope`
4. **A JWT** signed with the deployment's JWT secret, containing the username
   (`preferred_username`) and expiration based on `spec.ttl`. Roles are **not**
   put into the JWT. They are resolved from the user role bindings at
   evaluation time.
5. **A Kubernetes Secret** (owned by the token resource) holding the JWT under
   the `token` key.

## Retrieving the Token

Get the Secret name from the token status:

```bash
kubectl get arangopermissiontoken my-service-token -o jsonpath='{.status.secret.name}'
```

Then:

```bash
kubectl get secret <secret-name> -o jsonpath='{.data.token}' | base64 -d
```

## Attaching the Token to a Pod

The operator's mutating webhook can mount the token into pods of the
deployment. Add the label `permissions.arangodb.com/token: <token-name>` to the
Pod, or to its ServiceAccount. The pod also needs the label
`profiles.arangodb.com/deployment: <deployment-name>`. The token must be
`Ready` and belong to that deployment.

The webhook then:

- mounts the Secret read-only at `/var/run/secrets/arango/token/`
- sets the environment variable `ARANGO_TOKEN` to the token file path
  (`/var/run/secrets/arango/token/token`)
- labels the pod `permissions.arangodb.com/attached: <token-name>`

## Token Lifecycle

- Tokens are **automatically refreshed** at half their TTL.
- The JWT is **regenerated** when:
  - the signing secret changes
  - the TTL changes
  - the ArangoDB user changes
- Changes to `roles`, `policy` or `scope` do **not** require a new JWT. The
  operator updates the sidecar policy, role and bindings, and they take effect
  on the next evaluation (see [Runtime Behavior](behavior.md)).
- On **deletion**, the operator cleans up:
  - the ArangoDB user
  - the managed role and the managed policy in the sidecar

## Example: Read-Only Service Token

```yaml
apiVersion: permission.arangodb.com/v1alpha1
kind: ArangoPermissionToken
metadata:
  name: monitoring-token
spec:
  deployment:
    name: production
  ttl: 4h
  policy:
    statements:
      - effect: Allow
        actions:
          - "db:UseApiVersion"
          - "db:Read"
        resources:
          - "*"
      - effect: Deny
        actions:
          - "db:WriteData"
          - "db:WriteMeta"
          - "db:Create"
          - "db:Drop"
        resources:
          - "*"
  scope:
    statements:
      - effect: Allow
        actions:
          - "db:UseApiVersion"
          - "db:Read"
        resources:
          - "*"
```

This creates a token scoped to read-only access. The scope acts as a boundary:
even if the policy were later changed to grant broader permissions, this
token's managed role would stay limited to read operations.

## Example: Token with a Predefined Role

```yaml
apiVersion: permission.arangodb.com/v1alpha1
kind: ArangoPermissionToken
metadata:
  name: reports-reader
spec:
  deployment:
    name: production
  roles:
    - role:
        direct: managed:predefined:coredb-reader
      scope:
        policy:
          statements:
            - effect: Allow
              actions: ["*"]
              resources: ["db:apiversion:*", "db:database:reports", "db:*:reports:*"]
```
