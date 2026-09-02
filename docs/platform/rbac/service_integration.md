# Making a Service RBAC-Ready

Guide for ArangoDB platform service developers who want to integrate their
service with the RBAC system.

## Overview

A service becomes RBAC-ready by checking permissions before performing
actions. The sidecar handles authentication and authorization — your service
calls the sidecar's gRPC API to evaluate permissions.

## Prerequisites

- The service runs as a pod with the platform sidecar injected
  (label `profiles.arangodb.com/deployment: <deployment-name>`)
- The sidecar exposes the authorization API at `127.0.0.1:9201`

## Step 1: Get the Bearer Token

The user's JWT token arrives with the request — either in the `Authorization:
Bearer <token>` header (HTTP) or gRPC metadata. Your service does not need to
parse it or extract identity. Just forward it to the sidecar.

## Step 2: Evaluate Permissions with EvaluateToken

Call `EvaluateToken` on the authorization sidecar. The sidecar validates the
JWT, resolves the user's roles (from user bindings), and evaluates the
permission — all in one call.

```go
import pbAuthorizationV1 "github.com/arangodb/kube-arangodb/integrations/authorization/v1/definition"

client := pbAuthorizationV1.NewAuthorizationV1Client(sidecarConn)

resp, err := client.EvaluateToken(ctx, &pbAuthorizationV1.AuthorizationV1PermissionTokenRequest{
    Token:    bearerToken,
    Action:   "myservice:ReadData",
    Resource: "database:mydb",
})

if resp.GetEffect() == sidecarSvcAuthzTypes.Effect_Deny {
    // Return 403 Forbidden
}
```

For batch checks (multiple action/resource pairs in one call), use
`EvaluateTokenMany`:

```go
resp, err := client.EvaluateTokenMany(ctx, &pbAuthorizationV1.AuthorizationV1PermissionTokenManyRequest{
    Token: bearerToken,
    Items: []*pbAuthorizationV1.AuthorizationV1PermissionManyRequestItem{
        {Action: "myservice:ReadData", Resource: "database:mydb"},
        {Action: "myservice:WriteData", Resource: "database:mydb"},
    },
})
// resp.Items[0].GetEffect(), resp.Items[1].GetEffect()
```

Your service never needs to extract the user, resolve roles, or handle JWT
validation — `EvaluateToken` does it all.

### Action Naming Convention

Actions follow the `<namespace>:<verb>` pattern:

| Pattern | Example |
|---------|---------|
| `<service>:<action>` | `filestore:Read`, `filestore:Write` |
| `<service>:<resource>:<action>` | `database:collection:Create` |
| Wildcard | `filestore:*` (matches all filestore actions) |

### Resource Naming Convention

Resources identify what the action applies to:

| Pattern | Example |
|---------|---------|
| `<type>:<name>` | `database:mydb`, `collection:users` |
| `<type>:<path>` | `file:/data/reports/q1.csv` |
| Wildcard | `*` (all resources) |

## Step 3: Register Actions in Documentation

Document the actions your service checks so that administrators can create
appropriate policies. Add your actions to the service's API documentation.

Example for a file store service:

| Action | Description |
|--------|-------------|
| `filestore:Read` | Read files |
| `filestore:Write` | Create or update files |
| `filestore:Delete` | Delete files |
| `filestore:List` | List directory contents |

## Step 4: Handle the Superuser Case

Requests with no user identity (operator-internal JWT) are treated as
superuser by the sidecar — they bypass evaluation. Your service does not
need to handle this case; the sidecar returns `Effect_Allow` automatically.

## Step 5: Handle Evaluation Failures

If the sidecar is unreachable or returns an error:

- In `Central` mode: fail the request (return 503)
- In `CentralPermissive` mode: the sidecar handles this — it returns Allow
  on error. Your service just checks the response effect.

Do not implement your own fallback logic. The sidecar's mode determines the
behavior.

## Quick Checklist

- [ ] Pod has `profiles.arangodb.com/deployment` label
- [ ] Service connects to sidecar at `127.0.0.1:9201`
- [ ] Every protected endpoint calls `EvaluateToken()` with the bearer token
- [ ] Use `EvaluateTokenMany()` for batch permission checks
- [ ] Actions follow `<namespace>:<verb>` naming
- [ ] Resources follow `<type>:<name>` naming
- [ ] Actions are documented for administrators
- [ ] No custom fallback logic — trust the sidecar mode
- [ ] No JWT parsing — let the sidecar handle validation and role resolution
