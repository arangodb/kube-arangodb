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

The full catalog, with the intended access per role and its implementation
status, is described in the [design notes](../../../design/rbac/predefined_roles.md).
By default no access is granted (default deny); a role grants nothing until it is
bound to a user with a scope.

## Assigning a predefined role to a user

Bind a user to a predefined role with an `ArangoPermissionRoleUserBinding`. The
binding carries the **scope** that restricts where the role applies (resource-type
level, with one/many/pattern matching). Reference the predefined role by its
`managed:predefined:…` name:

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
    name: managed:predefined:coredb-reader
  scope:
    statements:
      - effect: Allow
        actions: ["*"]
        resources: ["reports", "reports/*"]
```

The `scope` above limits Alice's `coredb-reader` access to the `reports` database
and everything under it. To grant the role everywhere, use an Allow-all scope
(`actions: ["*"]`, `resources: ["*"]`).

You can also assign and scope roles through the management API (see
[User Role Bindings](user_bindings.md)):

```bash
curl -X PUT https://<gateway>/_management/permissions/user/alice/role/managed:predefined:coredb-reader \
  -H 'Content-Type: application/json' \
  -d '{"scope": {"statements": [{"effect": "Allow", "actions": ["*"], "resources": ["reports/*"]}]}}'
```

> `super-admin` is reserved and cannot be assigned - the operator binds it to the
> deployment `root` user automatically.

## Extending predefined roles

Predefined roles cannot be renamed or deleted, but you **can attach additional
policies** to them. Define an `ArangoPermissionPolicy` and bind it to a predefined
role by referencing the role's `managed:predefined:*` name in an
`ArangoPermissionPolicyRoleBinding` (predefined roles have no `ArangoPermissionRole`
CRD - they are referenced by their sidecar name directly). The operator merges the
bound policy into the predefined role, alongside its bundled policy.

For example, to grant everyone with `coredb-reader` write access on `reports`:

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
        actions: ["database:read", "database:write"]
        resources: ["reports/*"]
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
    name: managed:predefined:coredb-reader   # a predefined role - referenced by sidecar name
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
