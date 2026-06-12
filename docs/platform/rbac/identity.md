---
layout: page
title: Identity and Permissions
parent: RBAC
grand_parent: ArangoDBPlatform
nav_order: 5
---

# Identity and Permission Checks

The platform provides endpoints to inspect the current user's identity and
check permissions before performing actions.

## Gateway Endpoint Aliases

The gateway exposes short aliases for the most commonly used authentication
endpoints. These are public-facing routes that proxy to the internal sidecar
`/_integration/authn/v1/*` endpoints:

| Gateway alias | Proxied to                           |
|---------------|--------------------------------------|
| `/_identity`  | `/_integration/authn/v1/identity`    |
| `/_login`     | `/_integration/authn/v1/login`       |
| `/_logout`    | `/_integration/authn/v1/logout`      |

Both the short aliases and the full `/_integration/authn/v1/*` paths are
accessible through the gateway. The UI should use the short aliases.

## Endpoint Classification

| Endpoint | Audience | Purpose |
|---|---|---|
| `/_identity` | **UI** | Get current user identity |
| `/_login` | **UI** | Login with credentials |
| `/_logout` | **UI** | Logout, clear cookies |
| `/_management/permissions/validate` | **UI** | Check if current user can perform an action |
| `/_integration/authn/v1/validate` | Backend | Validate a JWT token programmatically |
| `/_integration/authn/v1/createToken` | Backend | Create JWT tokens for service accounts |
| `/_integration/authorization/v1/evaluate*` | Backend | Evaluate permissions for arbitrary user/roles |

For the initial UI release, only the **UI** endpoints are needed. The
`/_integration/*` endpoints are for backend/service-to-service use.

## Who Am I?

### GET `/_identity` (alias: `/_integration/authn/v1/identity`)

Returns the username of the currently authenticated user.

```bash
curl -H "Authorization: Bearer $TOKEN" \
  https://<gateway>/_integration/authn/v1/identity
```

Response:
```json
{
  "user": "alice"
}
```

If no user is specified in the token, `root` is returned.

### POST `/_integration/authn/v1/validate`

Returns detailed identity information including roles and token lifetime.

```bash
curl -X POST -H "Authorization: Bearer $TOKEN" \
  https://<gateway>/_integration/authn/v1/validate \
  -d '{"token": "<jwt-token>"}'
```

Response:
```json
{
  "is_valid": true,
  "message": "Token is valid",
  "details": {
    "lifetime": "3600s",
    "user": "alice",
    "roles": ["viewer", "managed:operator:abc-123"]
  }
}
```

## Can I?

### POST `/_management/permissions/validate`

Checks whether the **current authenticated user** can perform a specific
action on a resource. This is the self-check endpoint — it uses the identity
from the request context.

```bash
curl -X POST -H "Authorization: Bearer $TOKEN" \
  https://<gateway>/_management/permissions/validate \
  -d '{
    "action": "collection:write",
    "resource": "my-collection"
  }'
```

Response (allowed):
```json
{
  "message": "Access Granted",
  "effect": "Allow"
}
```

Response (denied):
```json
{
  "message": "Permission denied",
  "effect": "Deny"
}
```

### POST `/_integration/authorization/v1/evaluate`

Evaluates a permission for a given user and roles. This is the programmatic
endpoint used by integration services.

```bash
curl -X POST https://<gateway>/_integration/authorization/v1/evaluate \
  -d '{
    "user": "alice",
    "roles": ["viewer"],
    "action": "collection:read",
    "resource": "reports"
  }'
```

Response:
```json
{
  "message": "Access Granted",
  "effect": "Allow"
}
```

### POST `/_integration/authorization/v1/evaluate-many`

Batch evaluation — checks multiple action/resource pairs in a single request.

```bash
curl -X POST https://<gateway>/_integration/authorization/v1/evaluate-many \
  -d '{
    "user": "alice",
    "roles": ["viewer"],
    "items": [
      {"action": "collection:read", "resource": "reports"},
      {"action": "collection:write", "resource": "reports"},
      {"action": "database:read", "resource": "analytics"}
    ]
  }'
```

Response:
```json
{
  "message": "Access Granted",
  "effect": "Allow",
  "items": [
    {"message": "Access Granted", "effect": "Allow"},
    {"message": "Permission denied", "effect": "Deny"},
    {"message": "Access Granted", "effect": "Allow"}
  ]
}
```

The top-level `effect` is `Allow` only if all items are allowed.

### POST `/_integration/authorization/v1/evaluate-token`

Evaluates a permission by passing a JWT token directly. The user and roles
are extracted from the token claims.

```bash
curl -X POST https://<gateway>/_integration/authorization/v1/evaluate-token \
  -d '{
    "token": "<jwt-token>",
    "action": "collection:read",
    "resource": "reports"
  }'
```

### POST `/_integration/authorization/v1/evaluate-token-many`

Batch version of `evaluate-token`.

## Other Authentication Endpoints

### POST `/_login` (alias: `/_integration/authn/v1/login`)

Authenticates with username/password and returns a JWT token. Optionally
sets authentication cookies.

```bash
curl -X POST https://<gateway>/_integration/authn/v1/login \
  -d '{
    "credentials": {"username": "alice", "password": "secret"},
    "options": {"cookies": true}
  }'
```

### GET `/_logout` (alias: `/_integration/authn/v1/logout`)

Clears authentication cookies and redirects.

### POST `/_integration/authn/v1/createToken`

Creates a new JWT token for a specified user with given roles and lifetime.

```bash
curl -X POST https://<gateway>/_integration/authn/v1/createToken \
  -d '{
    "user": "alice",
    "roles": ["viewer", "editor"],
    "lifetime": "7200s"
  }'
```
