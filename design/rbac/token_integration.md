# RBAC Token Integration

## ArangoPermissionToken

The `ArangoPermissionToken` custom resource manages JWT tokens for accessing
ArangoDB deployments. It automates the creation of ArangoDB users, groups, and
signed JWT tokens.

## Token Spec

The token spec defines what the token grants access to:

- **`spec.roles[]`** — List of `ArangoPermissionScopedBindingRef` entries. Each
  entry binds an `ArangoPermissionRole` CRD with a **required scope boundary**
  (`ArangoPermissionScope`). The scope can be an inline policy or a reference to
  an `ArangoPermissionPolicy` CRD. Resolved role references are tracked in
  `status.roles`.
- **`spec.policy`** — Optional inline `Policy` definition. When set, the operator
  reconciles it as a dedicated sidecar policy and records the reference in
  `status.managedPolicy`.
- **`spec.scope`** — Boundary policy for the inline `spec.policy` on the token's
  managed group.
- **`spec.ttl`** — Token lifetime (default 1h, minimum 15m).

For fine-grained RBAC management, use the dedicated binding CRDs:

- **ArangoPermissionPolicy** — Define reusable policies
- **ArangoPermissionRole** — Define roles (policy containers)
- **ArangoPermissionPolicyRoleBinding** — Attach policies to roles
- **ArangoPermissionRoleUserBinding** — Assign roles to users with per-user scope

## Groups and the JWT

A **group** is the evaluation unit carried in JWTs. Each group bundles roles
(resolved to policies) with a required scope boundary. The JWT `groups` claim
carries group names (strings).

When the token handler reconciles `spec.roles`, each
`ArangoPermissionScopedBindingRef` produces a group in the sidecar. The scope
on each entry acts as the group's boundary — the token's permissions through
that role are restricted to what the scope allows.

The managed group (`managed:operator:<uid>`) is created when `spec.policy` is
set. It references the managed policy and uses `spec.scope` as its boundary.

### JWT Claims

The JWT contains:

- `user` — username string
- `groups` — array of group name strings
- Standard claims: `iss`, `exp`, `iat`

The JWT does **not** contain the scope. Scopes are stored server-side on the
group objects. The authorization service resolves groups to their scopes at
evaluation time. Changing a scope takes effect without re-issuing the JWT.

Example:

```json
{"user": "alice", "groups": ["managed:operator:abc123"]}
```

## Scope as Permission Boundary

Every group requires a scope. The scope acts as an **intersection boundary**:

1. **Policies** — Named policy references that define what permissions exist
2. **Scope** — A policy that limits what those policies can grant

A permission is granted only when **both** the policies allow it AND the scope
allows it. The scope cannot expand permissions — it can only restrict them.

The `ArangoPermissionScope` type supports two forms:
- **Inline** — `scope.policy` with statements directly
- **Reference** — `scope.ref` naming an `ArangoPermissionPolicy` CRD

Exactly one must be set (mutually exclusive).

**A group without a scope is skipped during evaluation.** To grant unrestricted
access through a group's policies, set the scope to allow all
(`effect: allow, actions: ["*"], resources: ["*"]`).

## CRD vs Database Storage

The RBAC system uses two storage layers:

- **CRDs (Kubernetes)** define the declarative configuration: policies
  (`ArangoPermissionPolicy`), roles (`ArangoPermissionRole`), and their bindings
  (`ArangoPermissionPolicyRoleBinding`, `ArangoPermissionRoleUserBinding`). These
  are the source of truth for operator-managed RBAC objects.

- **ArangoDB collections** (`_policies`, `_roles`, `_user_role_bindings`) are the
  runtime store used by the authorization sidecar for evaluation. The operator
  reconciles CRDs into these collections.

## Reconciliation Flow

The handler runs three sequential chains via `HandleP4`:

### Chain 1: HandleDeploymentConnection

1. **HandleArangoDBUser** — Creates an ArangoDB database user with a random
   password. Stores the user reference in `status.user`.

2. **HandleArangoSecret** — Creates a Kubernetes Secret and generates a signed
   JWT token containing:
   - Username from `status.user`
   - Group names from resolved `spec.roles` + the managed group name
     `managed:operator:<uid>` (if `status.role` is set)
   - Expiration based on `spec.ttl`

### Chain 2: HandleArangoDBPolicy

3. **HandleArangoDBPolicy** — No-op for direct policy resolution (removed).
   Inline policy via `spec.policy` is handled by `HandleManagedPolicy`.

### Chain 3: HandleDeploymentSidecarConnection

4. **HandleManagedPolicy** — If `spec.policy` is set, reconciles the inline
   policy into the sidecar and records the reference in `status.managedPolicy`.

5. **HandleArangoDBRole** — Creates a managed group in the sidecar
   (`managed:operator:<uid>`) that references all policy names from
   `status.managedPolicy`, and uses the rendered `spec.scope` as its scope.

## JWT Regeneration

The JWT includes a hash computed from:
- Username
- Signing secret hash
- TTL
- Group names list

When `HandleArangoDBRole` sets `status.role` (after creating the group), the
hash changes on the next reconcile cycle. `HandleArangoSecret` detects the
mismatch and regenerates the JWT with the managed group included.

**Important:** There is a one-cycle delay between group creation and the JWT
containing the managed group. Clients should wait for the `ReadyRole` condition
before using the token for RBAC-protected operations.

## Spec Hashing

Every permission CRD spec type exposes a `Hash()` method that computes a
deterministic SHA256 from its semantic fields. Named list types
(`ArangoPermissionScopedBindingRefList`, `ArangoPermissionBindingRefList`) also
expose `Hash()` for consistent hashing of collections.

After successful spec validation, the handler sets the `SpecAccepted` condition
with the hash via `UpdateWithHash`. This allows:

- **Change detection** — the operator only re-reconciles when the spec hash
  changes
- **Test observability** — tests wait for `SpecAccepted` condition hash to
  match `Spec.Hash()` to confirm the operator processed an update

## Finalizers

The token handler registers finalizers for cleanup:
- `FinalizerArangoPermissionUser` — Deletes the ArangoDB user
- `FinalizerArangoPermissionRole` — Deletes the group from the sidecar
- `FinalizerArangoPermissionPolicy` — Deletes the managed policy from the sidecar

## Key Files

- `pkg/apis/permission/v1alpha1/token_spec.go` — Token spec (deployment, roles, ttl, policy, scope)
- `pkg/apis/permission/v1alpha1/binding_ref.go` — `ArangoPermissionScopedBindingRef`, `ArangoPermissionScope`, list types with `Hash()`
- `pkg/handlers/permission/token/handler.go` — Main handler, JWT generation
- `pkg/handlers/permission/token/handler_role.go` — Group CRUD in sidecar
- `pkg/handlers/permission/token/handler_managed_policy.go` — Inline policy reconciliation
