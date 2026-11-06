---
layout: page
title: User Role Bindings
parent: RBAC
grand_parent: ArangoDBPlatform
nav_order: 5
---

# User Role Bindings

User role bindings assign roles to specific users with per-user scope policies.
This allows the same role to grant different permissions depending on who holds
it.

## API Endpoints

All endpoints require authentication and the appropriate RBAC permission.

### List bindings for a user

```bash
curl -H "Authorization: Bearer $TOKEN" https://<gateway>/_management/permissions/user/alice/role
```

Response:
```json
{
  "bindings": [
    {
      "role": "editor",
      "scope": {
        "statements": [
          {
            "effect": "Allow",
            "actions": ["*"],
            "resources": ["db:apiversion:*", "db:database:reports", "db:collection:reports:*"]
          }
        ]
      }
    }
  ]
}
```

### Assign a role to a user

```bash
curl -X POST -H "Authorization: Bearer $TOKEN" https://<gateway>/_management/permissions/user/alice/role/editor \
  -d '{
    "scope": {
      "statements": [
        {
          "effect": "Allow",
          "actions": ["*"],
          "resources": ["db:apiversion:*", "db:database:reports", "db:collection:reports:*"]
        }
      ]
    }
  }'
```

The `scope` is required and defines the inline policy for this specific
binding. Use `POST` to create a binding and `PUT` to replace the scope of an
existing one; `PUT` on a binding that doesn't exist returns `404`.

### Remove a role from a user

```bash
curl -X DELETE -H "Authorization: Bearer $TOKEN" https://<gateway>/_management/permissions/user/alice/role/editor
```

### Replace scope for a binding

```bash
curl -X PUT -H "Authorization: Bearer $TOKEN" https://<gateway>/_management/permissions/user/alice/role/editor \
  -d '{
    "scope": {
      "statements": [
        {
          "effect": "Allow",
          "actions": ["db:UseApiVersion", "db:Read"],
          "resources": ["*"]
        }
      ]
    }
  }'
```

## Required RBAC Permissions

| Operation | RBAC Action |
|---|---|
| List bindings | `rbac:ListUserRoleBinding` |
| Assign role | `rbac:AssignUserRole` |
| Remove role | `rbac:RemoveUserRole` |
| Replace scope | `rbac:ReplaceUserRoleScope` |

## How Scoping Works

A user's permissions are evaluated only against the roles bound to that user
(via an `ArangoPermissionRoleUserBinding` or a token). For each bound role two
things are combined:

1. **Role's named policies** - The policies attached to the role
2. **Binding's scope** - Inline policy specific to this user-role assignment

The effective permission is the **intersection**: an action is granted only when
the role's policies allow it **and** the binding scope allows it. Roles do not
carry a scope of their own. Evaluation uses the standard deny-by-default
algorithm, so a binding scope can further **restrict** access (an explicit Deny,
or simply not allowing an action) while never granting beyond the role's policies.

### Example: Same Role, Different Scopes

The `editor` role grants `db:UseApiVersion`, `db:Read` and `db:WriteData`
on all databases and collections. But each user gets a different scope:

**Alice** - Can edit only the collections of the `reports` database:
```json
{
  "scope": {
    "statements": [
      { "effect": "Allow", "actions": ["*"], "resources": ["db:apiversion:*", "db:database:reports", "db:collection:reports:*"] }
    ]
  }
}
```

**Bob** - Can edit only the collections of the `analytics` database:
```json
{
  "scope": {
    "statements": [
      { "effect": "Allow", "actions": ["*"], "resources": ["db:apiversion:*", "db:database:analytics", "db:collection:analytics:*"] }
    ]
  }
}
```
