---
layout: page
title: Group Role Bindings
parent: RBAC
grand_parent: ArangoDBPlatform
nav_order: 6
---

# Group Role Bindings

Group role bindings assign a role (with a scope policy) to a **group** instead of
an individual user. Every request whose identity token lists the group in its
`groups` claim inherits the role. This complements per-user bindings and lets you
manage access by team or directory group.

A request's effective permissions are the **union** of:

1. the roles bound directly to the user, and
2. the roles bound to **every** group carried in the token's `groups` claim.

Group bindings are stored in a dedicated pool keyed `group:role`, independent of
user bindings, so a user whose name happens to start with `group:` never inherits
a group's roles - only the token's `groups` claim grants group-derived access.

## API Endpoints

All endpoints require authentication and the appropriate RBAC permission.

### List bindings for a group

```bash
curl https://<gateway>/_management/permissions/group/developers/role
```

### Assign a role to a group

```bash
curl -X POST https://<gateway>/_management/permissions/group/developers/role/editor \
  -d '{
    "scope": {
      "statements": [
        { "effect": "Allow", "actions": ["collection:*"], "resources": ["reports"] }
      ]
    }
  }'
```

The `scope` is required and defines the inline policy for this specific binding.

### Remove a role from a group

```bash
curl -X DELETE https://<gateway>/_management/permissions/group/developers/role/editor
```

### Replace scope for a binding

```bash
curl -X PUT https://<gateway>/_management/permissions/group/developers/role/editor \
  -d '{
    "scope": {
      "statements": [
        { "effect": "Allow", "actions": ["collection:read"], "resources": ["*"] }
      ]
    }
  }'
```

## Required RBAC Permissions

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

With an `editor` role bound to `developers`, `alice` gains that role because her
token lists `developers` - in addition to any roles bound to `alice` directly.

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
