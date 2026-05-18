# RBAC Architecture

## Request Flow

```
 HTTP Request
      |
      v
+----------+    CheckRequest    +----------------+
|  Envoy   | ----------------> | ext_authz      |
|  Gateway |                    | Handler Chain  |
+----------+                    +-------+--------+
                                        |
                                        v
                               +-----------------+
                               | gRPC Interceptor|
                               | (extract JWT)   |
                               +-------+---------+
                                        |
                                        v
                               +-----------------+
                               | Authorization   |
                               | Plugin          |
                               +-------+---------+
                                        |
                                        v
                               +-----------------+
                               | Policy          |
                               | Evaluation      |
                               +-----------------+
```

**Stage descriptions:**

1. **HTTP Request → Envoy Gateway** — At this stage: user identity is not yet validated. The raw HTTP request (with optional Bearer token or cookie) arrives at the Envoy proxy.
2. **Envoy → ext_authz Handler Chain** — At this stage: Envoy delegates the `CheckRequest` to the ext_authz service. The handler chain inspects headers and cookies but authentication is still in progress.
3. **gRPC Interceptor (extract JWT)** — At this stage: JWT is validated, user identity (username + groups) is extracted and attached to the gRPC context. Invalid tokens are rejected here.
4. **Authorization Plugin** — At this stage: the authenticated identity is known. The plugin chain performs the permission check against the configured mode (Central, CentralPermissive, etc.).
5. **Policy Evaluation** — At this stage: all applicable policies are collected and evaluated. The final Allow/Deny decision is returned.

## Envoy ext_authz

HTTP requests hit the Envoy gateway, which sends a `CheckRequest` to the
ext_authz gRPC service. A handler chain processes the request through
authentication stages: bearer token, cookie, OpenID Connect, session handling,
and pass-mode configuration. Each route can override auth behavior via
`ExtAuthzPerRoute` context extensions.

Key files:
- `integrations/envoy/auth/v3/impl.go` - `Check()` RPC implementation
- `integrations/envoy/auth/v3/shared/handler.go` - `AuthHandler` interface
- `integrations/envoy/auth/v3/shared/factory.go` - Handler chain factory

## gRPC Interceptor

The interceptor validates the JWT token from the request metadata, extracts the
user identity (username + groups list), and attaches it to the gRPC context.
Subsequent service methods retrieve the identity via
`authenticator.GetIdentity(ctx)`.

Specifically the interceptor:
- **Modifies the gRPC context** — it attaches the extracted `Identity` (username
  and groups) to the context so downstream handlers can access it via
  `authenticator.GetIdentity(ctx)`.
- **Rejects invalid tokens** — requests with a malformed or expired JWT are
  immediately rejected with an `Unauthenticated` gRPC error. The request never
  reaches the authorization layer.
- **Handles missing identity** — when the token is valid but carries no user
  claim (e.g. internal service-to-service calls), the identity is set to `nil`.
  A nil-user is treated as a superuser and bypasses authorization checks.

Key files:
- `pkg/util/svc/authenticator/interceptor.go` - Unary and stream interceptors
- `pkg/util/svc/authenticator/identity.go` - `Identity` struct with `EvaluatePermission()`
- `pkg/util/svc/authenticator/jwt.go` - JWT validation and identity extraction

## Authorization Plugin Chain

The plugin chain determines how authorization decisions are made:

1. **Superuser check** - Nil user (service accounts) always granted
2. **Plugin dispatch** - Routes to the configured plugin mode:
   - `Central` - Strict policy evaluation
   - `CentralPermissive` - Wraps Central, converts errors to Allow
   - `Always` / `Never` - Bypass modes
3. **Policy evaluation** - Collects and evaluates all applicable policies

Key files:
- `integrations/authorization/v1/shared/authz.go` - `Plugin` interface
- `integrations/authorization/v1/shared/superuser.go` - Superuser bypass
- `integrations/authorization/v1/shared/permissive.go` - Error-tolerant wrapper
- `integrations/authorization/v1/configuration.go` - Configuration types

## Feature Flags

RBAC is controlled by two feature flags that determine the authorization mode
at runtime:

| Flag | Default | Effect |
|---|---|---|
| `central-services` | Disabled | Enables authorization sidecar service |
| `rbac-enforced` | Disabled | Switches from permissive to strict enforcement |

Feature flags are set as operator CLI arguments and are typically passed via Helm
values or the operator Deployment spec. For example:

```
--deployment.feature.central-services=true
--deployment.feature.rbac-enforced=true
```

See [enabling.md](enabling.md) for the full enablement procedure.

Both depend on `gateway-integration`. The mode is resolved in
`pkg/integrations/sidecar/integration.authorization.v1.go` and passed as
`INTEGRATION_AUTHORIZATION_V1_TYPE` env var.

Key files:
- `pkg/deployment/features/rbac.go` - `rbac-enforced` feature definition
- `pkg/deployment/features/sidecar.go` - `central-services` feature definition
- `pkg/integrations/sidecar/integration.authorization.v1.go` - Mode resolution
- `pkg/deployment/resources/internal_sidecar.go` - Sidecar args injection

## Propagation Delay

When groups or policies are modified, changes propagate within the pooler refresh
interval (10 s by default) plus the client sync interval (15 s). Users do **not**
need to re-login after a policy change; existing JWTs remain valid, but the
authorization data they map to is refreshed server-side on the next pooler cycle.

## Degraded Mode Behavior

If the authorization sidecar becomes unreachable, the behavior depends on the
configured mode:

| Mode | Behavior when sidecar is unreachable |
|---|---|
| `Central` | All authorization requests return **Deny**. |
| `CentralPermissive` | All authorization requests return **Allow** and a warning is logged for every permitted request. |

Internal service-to-service calls that carry a nil user (no identity claim in the
JWT) bypass the authorization sidecar entirely, so they are unaffected by sidecar
availability.
