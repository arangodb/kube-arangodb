# RBAC Policy Evaluation

## Core Concepts

- **Policy** — A set of statements (effect + action patterns + resource patterns)
- **Role** — A named container of policy references, managed via CRDs
- **Group** — The evaluation unit carried in JWTs. A group bundles roles
  (resolved to policies) with a required **scope boundary**. Groups are stored
  in the sidecar's role pool and referenced by name.
- **Scope** — An inline policy or policy CRD reference that acts as an
  intersection boundary on a group. A permission is granted only when the
  policies allow it AND the scope allows it.

## Evaluation Algorithm

The evaluation follows a **deny-by-default** model with **groups**.

The JWT carries group names (strings). Each group is resolved to a
`ScopedPolicy` containing its policies and scope. Evaluation proceeds per
group:

1. Collect all matching groups for the requested group names.
2. For each group, resolve policies + scope. Groups without a scope are
   **skipped entirely** — they do not contribute to the result.
3. For each group:
   a. Evaluate the **policies** first:
      - If a statement matches the action/resource AND its effect is **Deny**,
        return Deny immediately (explicit deny wins).
      - If a statement matches AND its effect is **Allow**, mark as allowed.
      - If no statement matches, the policy denies by default.
   b. If the policies **denied**, skip to the next group.
   c. If the policies **allowed**, evaluate the **scope boundary**:
      - The scope is evaluated with the same statement-matching logic.
      - If the scope **denies** the action, the group's grant is discarded.
      - If the scope **allows** the action, access is granted.
4. If any group granted access (policies allow AND scope allows), return Allow.
5. Otherwise, return Deny (implicit deny).

The scope acts as an **intersection boundary**: an action is only allowed when
**both** the policies and the scope agree.

### Statement Matching

A statement matches when both conditions are true:
- At least one action pattern matches the requested action
- At least one resource pattern matches the requested resource

Pattern matching supports colon-separated segments. For example, action
`rbac:ListRole` is split into `["rbac", "ListRole"]` and matched segment by
segment against the pattern.

Key files:
- `pkg/sidecar/services/authorization/client/evaluate.go` — `EvaluatePolicies()`
- `pkg/sidecar/services/authorization/client/scope.go` — `ScopedPolicy`, `ScopedPolicies` (map), `ScopedPolicy.Evaluate()`, `ScopedPolicies.Evaluate()`
- `pkg/sidecar/services/authorization/client/policy.go` — `Policy.Evaluate()`, `Statement.Evaluate()`

## Group Resolution

Groups are resolved from two paths:

### Client-side (streaming cache)

Used by the integration service when connecting to the authorization sidecar
over gRPC:

1. The `internalCache` maintains roles, policies, and user role bindings.
2. `extractGroups(user, groupNames...)` resolves groups from both explicit
   names and user bindings (`<user>:<group>` keys), builds a `ScopedPolicy`
   per group (skipping groups with nil scope), and returns a `ScopedPolicies` map.

Key files:
- `pkg/sidecar/services/authorization/client/cache.go`
- `pkg/sidecar/services/authorization/client/client.go`

### Server-side (direct pool access)

Used by the authorization sidecar itself:

1. `getUserGroups(user, groupNames...)` resolves groups from both explicit
   names and user bindings, builds a `ScopedPolicy` per group (skipping
   groups with nil scope), and returns a `ScopedPolicies` map.

Key file:
- `pkg/sidecar/services/authorization/client.go`

## Group Management APIs

The sidecar exposes CRUD APIs for managing groups:

| API | Permission | Description |
|-----|-----------|-------------|
| `APIListGroup` | `rbac:ListGroup` | List all groups |
| `APIGetGroup` | `rbac:GetGroup` | Get group by name |
| `APICreateGroup` | `rbac:CreateGroup` | Create a new group |
| `APIUpdateGroup` | `rbac:UpdateGroup` | Update an existing group |
| `APIDeleteGroup` | `rbac:DeleteGroup` | Delete a group |

Key file:
- `pkg/sidecar/services/authorization/impl_api_groups.go`

## Configuration Modes

The authorization plugin mode is set per deployment:

| Mode | Behavior |
|---|---|
| `Central` | Strict enforcement. Deny on evaluation failure. |
| `CentralPermissive` | Central with fallback to Allow on error. Logs the error. |
| `Always` | Always returns Allow. Used for development/testing. |
| `Never` | Always returns Deny. Locks out all access. |

The mode is configured per deployment via the `--sidecar.auth.mode` CLI flag on
the authorization sidecar container. The value is determined by the
`rbac-enforced` feature flag:
- Enabled → `Central` (strict)
- Disabled → `CentralPermissive` (permissive, default)

The mode cannot be changed at runtime — changing it requires a pod restart
(the operator updates the sidecar container args and the pod is recreated).

> **Warning:** The default `CentralPermissive` mode should NOT be changed to
> `Central` unless RBAC policies are properly configured and tested. Switching
> to strict enforcement without adequate policies in place will deny all
> non-superuser requests.

Key file:
- `pkg/deployment/resources/internal_sidecar.go`

### Superuser Bypass

Requests with no user and no groups are treated as superuser (operator-internal
JWT) and bypass policy evaluation entirely.

> **NOTE:** Nil users (superuser identity) cannot be created from outside the
> system. The authentication layer (gRPC interceptor, ext_authz handler) always
> extracts a username from valid JWTs. Only internal service-to-service calls
> that bypass authentication produce nil-user identities.

### Health Check Bypass

gRPC health check endpoints (`/grpc.health.v1.Health/`) are excluded from
authentication to allow proxy upstream health checks and Kubernetes probes.

Key file:
- `pkg/util/svc/authenticator/interceptor.go`
