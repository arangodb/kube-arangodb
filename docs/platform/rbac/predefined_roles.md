---
layout: page
title: Predefined Roles
parent: RBAC
grand_parent: ArangoDBPlatform
nav_order: 2
---

# Predefined Roles

When RBAC is enabled, the operator automatically creates a catalog of
**predefined roles** in every deployment's authorization sidecar. You assign and
scope these roles to users; you cannot edit or delete them (they are
operator-managed and visible but not editable).

Predefined roles are named under the reserved prefix `managed:predefined:`, e.g.
`managed:predefined:coredb-reader`.

| Role | Purpose |
|---|---|
| `managed:predefined:super-admin` | Full access. Reserved - bound automatically to the `root` user and **not assignable** by customers. |
| `managed:predefined:tenant-admin` | Manages users and role bindings. |
| `managed:predefined:coredb-reader` | Read-only database operations on scoped resources. |
| `managed:predefined:coredb-developer` | Read and write database operations on scoped resources. |
| `managed:predefined:coredb-admin` | Manages scoped resources' structures and lifecycle. |
| `managed:predefined:ai-user` | Executes AI workflows and reads outputs on scoped resources. |
| `managed:predefined:ai-developer` | Builds, configures, manages, and executes AI workflows on scoped resources. |
| `managed:predefined:platform-operator` | Operates platform and bundled services, views observability, starts containers. |
| `managed:predefined:secret-admin` | Manages secrets on scoped resources. |

By default no access is granted (default deny). A role grants nothing until it is
bound to a user with a scope.

> **Note:** The operator only ships a bundled policy for `super-admin`
> (Allow-all). It creates all other predefined roles **without policies**. The
> permissions are added separately by attaching policies to the roles, as
> described in [Extending predefined roles](#extending-predefined-roles).
>
> The platform components ship these policies in their Helm charts, as one
> `ArangoPermissionPolicy` plus one `ArangoPermissionPolicyRoleBinding` per role,
> for the ArangoDeployment set in `arangodb_platform.deployment.name`:
>
> | Roles | Policies provided by |
> |---|---|
> | `tenant-admin`, `coredb-reader`, `coredb-developer`, `coredb-admin` | [`arangodb-core-rbac-permissions`](https://github.com/arangodb/arangodb-core-rbac-permissions/blob/main/charts/arangodb-core-rbac-permissions/templates/predefined_roles_coredb.yaml) chart (CoreDB `db:*` actions) |
> | `ai-user`, `ai-developer`, `secret-admin` | `arango-control-plane` chart (AI services and Secrets Manager `ai:*` actions) |
>
> Until such policies are attached, binding a user to a predefined role grants
> nothing.

## Assigning a predefined role to a user

Bind a user to a predefined role with an `ArangoPermissionRoleUserBinding`. The
binding carries the **scope** that restricts where the role applies (resource-type
level, with one/many/pattern matching). Predefined roles have no `ArangoPermissionRole`
CRD, so reference them with the `direct` field (their exact sidecar name) rather than
`name` (which resolves an `ArangoPermissionRole` CRD):

```yaml
apiVersion: permission.arangodb.com/v1alpha1
kind: ArangoPermissionRoleUserBinding
metadata:
  name: alice-coredb-reader
spec:
  deployment:
    name: my-deployment
  userName: alice
  role:
    direct: managed:predefined:coredb-reader   # direct sidecar name (no CRD)
  scope:
    statements:
      - effect: Allow
        actions: ["*"]
        resources: ["db:apiversion:*", "db:database:reports", "db:*:reports:*"]
```

The `scope` above limits Alice's `coredb-reader` access to the `reports` database
and everything in it (`db:*:reports:*` covers its collections, views, analyzers and
graphs). `db:apiversion:*` is needed because ArangoDB checks `db:UseApiVersion` on
every request. To grant the role everywhere, use an Allow-all scope
(`actions: ["*"]`, `resources: ["*"]`).

You can also assign and scope roles through the management API (see
[User Role Bindings](user_bindings.md)):

```bash
curl -X POST -H "Authorization: Bearer $TOKEN" https://<gateway>/_management/permissions/user/alice/role/managed:predefined:coredb-reader \
  -H 'Content-Type: application/json' \
  -d '{"scope": {"statements": [{"effect": "Allow", "actions": ["*"], "resources": ["db:apiversion:*", "db:database:reports", "db:*:reports:*"]}]}}'
```

> `super-admin` is reserved: the operator binds it to the deployment `root` user
> automatically and **rejects** any `ArangoPermissionRoleUserBinding` or
> `ArangoPermissionPolicyRoleBinding` that references it (the binding is marked
> `SpecValid: false` and never applied), so customers cannot assign it or attach
> policies to it.

## Extending predefined roles

Predefined roles cannot be renamed or deleted, but you **can attach additional
policies** to them. Define an `ArangoPermissionPolicy` and bind it to a predefined
role by referencing the role's exact sidecar name through the `direct` field of an
`ArangoPermissionPolicyRoleBinding` (predefined roles have no `ArangoPermissionRole`
CRD, so `direct` is used instead of the CRD-resolving `name` field). The operator
merges the bound policy into the predefined role, alongside its bundled policy.

For example, to grant everyone with `coredb-reader` write access to the
collections of the `reports` database:

```yaml
# 1. Define the extra policy
apiVersion: permission.arangodb.com/v1alpha1
kind: ArangoPermissionPolicy
metadata:
  name: reports-writer
spec:
  deployment:
    name: my-deployment
  policy:
    statements:
      - effect: Allow
        actions: ["db:WriteData"]
        resources: ["db:collection:reports:*"]
---
# 2. Attach it to the predefined role by its managed:predefined: name
apiVersion: permission.arangodb.com/v1alpha1
kind: ArangoPermissionPolicyRoleBinding
metadata:
  name: coredb-reader-reports-writer
spec:
  deployment:
    name: my-deployment
  role:
    direct: managed:predefined:coredb-reader   # a predefined role - referenced directly by name
  policy:
    name: reports-writer
```

Every user bound to `coredb-reader` now additionally carries the `reports-writer`
policy. Deleting the binding removes the extra policy again.

Notes:

- The attached policy extends the role for **everyone** assigned it (it is not
  per-user). The operator reconciles the change within a short interval.
- `super-admin` already grants everything, so attaching policies to it has no
  effect.
- The predefined role's own bundled policy is preserved - bound policies are
  merged in, not replaced, and the operator does not clobber them on sync.

See [Policies and Roles](policies.md) for the object model and the action /
resource formats.

### Per-user extension (composition)

To grant a **single user** more than a predefined role provides (rather than
extending the role for everyone), bind them to an additional custom role. A user's
effective permissions are the **union of all their role bindings**, with `Deny`
taking precedence. Create a custom `ArangoPermissionPolicy` + `ArangoPermissionRole`
+ `ArangoPermissionPolicyRoleBinding`, then add a second
`ArangoPermissionRoleUserBinding` for that user - the user keeps their predefined
role and gains the custom one. Removing the extra binding reverts them.

## Errors on insufficient permissions

Services do not proactively hide unavailable actions; instead they return an
actionable error when an unauthorized action is attempted. If a user is denied,
verify that a role is bound to them and that the binding's **scope** covers the
target resource.
