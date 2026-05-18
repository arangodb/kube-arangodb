# RBAC Authorization

> **Alpha Feature** - RBAC is currently in alpha (`v1alpha1`). APIs, CRD schemas,
> and behavior may change in future releases without notice.

The operator provides a policy-based Role-Based Access Control (RBAC) system
for authorizing requests to ArangoDB deployments. Authorization is enforced by
the gateway sidecar through Envoy's ext_authz filter and gRPC interceptors.

## Sections

- [Architecture](architecture.md) - Request flow, Envoy ext_authz, gRPC interceptor, plugin chain
- [Data Model](data_model.md) - Policies, Roles, Scopes, UserRoleBindings, and CRD bindings
- [Evaluation](evaluation.md) - Policy evaluation logic and configuration modes
- [API](api.md) - Management API, Integration API (identity, canI, evaluate), proto definitions
- [Token Integration](token_integration.md) - ArangoPermissionToken reconciliation flow
- [Storage](storage.md) - ArangoDB collections, pooling, and sync

## File Inventory

### Proto Types (`pkg/sidecar/services/authorization/types/`)

| File | Type |
|---|---|
| `policy.proto` / `policy.go` | Policy, PolicyStatement, Effect |
| `role.proto` / `role.go` | Role (policies + scope) |
| `user_role_binding.proto` / `user_role_binding.go` | UserRoleBinding (role + scope) |
| `context.proto` | Context (parameter map for requests) |
| `effect.proto` | Effect enum (Allow, Deny) |

### Sidecar API (`pkg/sidecar/services/authorization/`)

| File | Purpose |
|---|---|
| `definition/api.proto` | AuthorizationAPI service (policy/role/binding CRUD + validate) |
| `definition/pool.proto` | AuthorizationPoolService (streaming sync) |
| `impl.go` | Service wiring, pooler init, health checks |
| `impl_validate.go` | ValidateSelfPermission |
| `impl_api_policy.go` | Policy CRUD endpoints |
| `impl_api_roles.go` | Role CRUD endpoints |
| `impl_api_user_role_bindings.go` | UserRoleBinding CRUD endpoints |
| `client.go` | Server-side policy evaluation (direct pool) |
| `client/client.go` | Client-side streaming cache |
| `client/cache.go` | Cache with role-to-policy maps |
| `client/evaluate.go` | `EvaluatePolicies()` |
| `client/policy.go` | Policy/Statement parsing and matching |

### Integration Services (`integrations/`)

| File | Purpose |
|---|---|
| `authentication/v1/definition/definition.proto` | AuthenticationV1 (identity, validate, login, logout, createToken) |
| `authentication/v1/implementation.go` | Authentication implementation |
| `authorization/v1/definition/definition.proto` | AuthorizationV1 (evaluate, evaluateMany, evaluateToken) |
| `authorization/v1/definition/request.proto` | Permission request/response messages |
| `authorization/v1/impl.go` | Authorization evaluation implementation |
| `authorization/v1/configuration.go` | ConfigurationType (Central, CentralPermissive, Always, Never) |
| `authorization/v1/shared/authz.go` | Plugin interface |
| `authorization/v1/shared/superuser.go` | Superuser bypass |
| `authorization/v1/shared/permissive.go` | Error-tolerant wrapper |
| `authorization/v1/shared/always.go` / `never.go` | Bypass plugins |

### CRDs (`pkg/apis/permission/`)

| File | CRD |
|---|---|
| `definitions.go` | Resource names and constants for all permission CRDs |
| `v1alpha1/policy.go` / `policy_spec.go` / `policy_status.go` | ArangoPermissionPolicy |
| `v1alpha1/role.go` / `role_spec.go` / `role_status.go` | ArangoPermissionRole |
| `v1alpha1/token.go` / `token_spec.go` / `token_status.go` | ArangoPermissionToken |
| `v1alpha1/policy_role_binding.go` / `_spec.go` / `_status.go` | ArangoPermissionPolicyRoleBinding |
| `v1alpha1/role_user_binding.go` / `_spec.go` / `_status.go` | ArangoPermissionRoleUserBinding |
| `v1alpha1/register.go` | Type registration with Kubernetes scheme |
| `v1alpha1/conditions.go` | Condition type constants |
| `v1alpha1/zz_generated.deepcopy.go` | Generated DeepCopy methods |

### PolicyRoleBinding Handler (`pkg/handlers/permission/policy_role_binding/`)

| File | Purpose |
|---|---|
| `handler.go` | Reconciliation loop, finalizer management |
| `handler_references.go` | Resolves Policy and Role CRD references |

### Token Handler (`pkg/handlers/permission/token/`)

| File | Purpose |
|---|---|
| `handler.go` | Main reconciliation loop, JWT generation |
| `handler_policy.go` | ArangoPermissionPolicy CRD resolution |
| `handler_role.go` | Managed role CRUD, `renderRole()`, `renderPolicy()` |

### Spec Hashing (`pkg/apis/permission/v1alpha1/`)

Each spec type (`*Spec`) has a `Hash() string` method:
- Computed from semantic fields (deployment name, policy content, binding refs, etc.)
- Includes `Description` fields
- Used by handlers to set `SpecAccepted` condition with `UpdateWithHash`
- Tests wait for condition hash to match `Spec.Hash()` to confirm processing

### Feature Flags (`pkg/deployment/features/`)

| File | Feature |
|---|---|
| `rbac.go` | `rbac-enforced` feature flag |
| `sidecar.go` | `central-services` feature flag |
