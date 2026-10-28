---
layout: page
title: Group Role Bindings
parent: RBAC
grand_parent: ArangoDBPlatform
nav_order: 6
---

# Group Role Bindings

Group role bindings assign a role (with a scope policy) to a **group** rather than
to an individual user. Every user whose identity token lists the group inherits
the role, with the per-binding scope applied. This lets you manage access by team
or directory group instead of per user.

Group membership is taken from the `groups` claim of the identity (JWT) token.
When a request is authorized, the effective permissions are the **union** of:

1. the roles bound directly to the user (via `ArangoPermissionRoleUserBinding`), and
2. the roles bound to **every** group listed in the token's `groups` claim (via
   `ArangoPermissionRoleGroupBinding`).

Group and user bindings are stored independently, so a user named the same as a
group never inherits that group's roles - only the token's `groups` claim grants
group-derived access.

## Defining a binding (CRD)

An `ArangoPermissionRoleGroupBinding` binds a role to a group within a
deployment:

```yaml
apiVersion: permission.arangodb.com/v1alpha1
kind: ArangoPermissionRoleGroupBinding
metadata:
  name: developers-editor
spec:
  deployment:
    name: my-deployment
  groupName: developers
  role:
    name: editor          # ArangoPermissionRole name, or `direct` for a predefined role
  scope:
    statements:
      - effect: Allow
        actions: ["collection:*"]
        resources: ["reports"]
```

- `groupName` is matched case-sensitively against the entries of the token's
  `groups` claim.
- `role` references an `ArangoPermissionRole` by name, or a predefined role by
  its direct sidecar name (for example `managed:predefined:coredb-reader`). The
  reserved `super-admin` role cannot be bound.
- `scope` is required and defines the inline policy for this binding.

## Management API

The operator reconciles the CRD into the authorization sidecar, which also exposes
the bindings directly (the same surface used without a CRD):

```bash
# List bindings for a group
curl https://<gateway>/_management/permissions/group/developers/role

# Assign a role to a group
curl -X POST https://<gateway>/_management/permissions/group/developers/role/editor \
  -d '{ "scope": { "statements": [ { "effect": "Allow", "actions": ["collection:*"], "resources": ["reports"] } ] } }'

# Replace scope / remove
curl -X PUT    https://<gateway>/_management/permissions/group/developers/role/editor -d '{ "scope": { ... } }'
curl -X DELETE https://<gateway>/_management/permissions/group/developers/role/editor
```

| Operation | RBAC Action |
|---|---|
| List bindings | `rbac:ListGroupRoleBinding` |
| Assign role | `rbac:AssignGroupRole` |
| Remove role | `rbac:RemoveGroupRole` |
| Replace scope | `rbac:ReplaceGroupRoleScope` |

## Identity token

Group membership comes from the `groups` claim of the token issued for the
request:

```json
{
  "preferred_username": "alice",
  "groups": ["developers", "oncall"]
}
```

With the binding above, `alice` gains the `editor` role scoped to `reports`
because her token lists `developers` - in addition to any roles bound to `alice`
directly.

## How Scoping Works

Group bindings scope exactly like user bindings. For each bound role the
effective permission is the **intersection** of:

1. **Role's named policies** - the policies attached to the role, and
2. **Binding's scope** - the inline policy on this group-role binding.

An action is granted only when the role's policies allow it **and** the binding
scope allows it. Across all resolved bindings (user and group), access is the
**union**: a user is allowed an action if any of their resolved role bindings
grants it under the deny-by-default algorithm. Group bindings therefore augment -
never restrict - a user's direct bindings.

### Example: team-wide read, personal write

Bind the whole `developers` group to a read-only scope of a role, and give an
individual a broader binding:

**Group `developers`** - read any collection:
```yaml
groupName: developers
role: { name: editor }
scope:
  statements:
    - { effect: Allow, actions: ["collection:read"], resources: ["*"] }
```

**User `alice`** - additionally edit the `reports` collection:
```yaml
userName: alice
role: { name: editor }
scope:
  statements:
    - { effect: Allow, actions: ["collection:*"], resources: ["reports"] }
```

A token for `alice` carrying `"groups": ["developers"]` resolves to both
bindings: she can read every collection (from the group binding) and write
`reports` (from her user binding).
