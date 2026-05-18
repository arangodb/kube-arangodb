# RBAC API

The RBAC system exposes two API surfaces: the **management API** (for managing
policies, roles, and bindings) and the **integration API** (for authentication
and permission evaluation).

## Management API

Exposed by the authorization sidecar under the `/_management/permissions/` prefix.
All endpoints require authentication and the appropriate RBAC permission.
Paths below are shown as full paths including the prefix.

### Permission Validation

| Method | Path | Description |
|---|---|---|
| `POST` | `/_management/permissions/validate` | Validate the caller's own permission for an action/resource |

Request body: `AuthorizationAPIValidateSelfRequest { action, resource, context }`
Response body: `AuthorizationAPIValidateResponse { message, effect }`

Implementation: `pkg/sidecar/services/authorization/impl_validate.go`
- Extracts identity from gRPC context via `authenticator.GetIdentity(ctx)`
- Calls `EvaluatePermission()` on the identity with the plugin chain
- No RBAC permission required (self-check only)

### Policy Management

| Method | Path | RBAC Action | Purpose |
|---|---|---|---|
| `GET` | `/_management/permissions/policy` | `rbac:ListPolicy` | List all policies |
| `GET` | `/_management/permissions/policy/{name}` | `rbac:GetPolicy` | Get a single policy by name |
| `POST` | `/_management/permissions/policy/{name}` | `rbac:CreatePolicy` | Create a new policy |
| `PUT` | `/_management/permissions/policy/{name}` | `rbac:UpdatePolicy` | Update an existing policy |
| `DELETE` | `/_management/permissions/policy/{name}` | `rbac:DeletePolicy` | Delete a policy |

Request body (POST/PUT): `AuthorizationAPIPolicyCreateRequest { rules[] }` where each rule contains `{ action, resource, effect }`
Response body (GET single): `AuthorizationAPIPolicyResponse { name, rules[] }`
Response body (GET list): `AuthorizationAPIPolicyListResponse { policies[] }`
Response body (POST/PUT): `AuthorizationAPIPolicyResponse { name, rules[] }`
Response body (DELETE): empty (204 No Content)

Implementation: `pkg/sidecar/services/authorization/impl_api_policy.go`

### Role Management

| Method | Path | RBAC Action | Purpose |
|---|---|---|---|
| `GET` | `/_management/permissions/role` | `rbac:ListRole` | List all roles |
| `GET` | `/_management/permissions/role/{name}` | `rbac:GetRole` | Get a single role by name |
| `POST` | `/_management/permissions/role/{name}` | `rbac:CreateRole` | Create a new role |
| `PUT` | `/_management/permissions/role/{name}` | `rbac:UpdateRole` | Update an existing role |
| `DELETE` | `/_management/permissions/role/{name}` | `rbac:DeleteRole` | Delete a role |

Request body (POST/PUT): `AuthorizationAPIRoleCreateRequest { policies[] }` where policies is a list of policy names bound to this role
Response body (GET single): `AuthorizationAPIRoleResponse { name, policies[] }`
Response body (GET list): `AuthorizationAPIRoleListResponse { roles[] }`
Response body (POST/PUT): `AuthorizationAPIRoleResponse { name, policies[] }`
Response body (DELETE): empty (204 No Content)

Implementation: `pkg/sidecar/services/authorization/impl_api_roles.go`

### User Role Binding Management

| Method | Path | RBAC Action | Purpose |
|---|---|---|---|
| `GET` | `/_management/permissions/user/{user}/role` | `rbac:ListUserRoleBinding` | List all role bindings for a user |
| `POST` | `/_management/permissions/user/{user}/role/{role}` | `rbac:AssignUserRole` | Assign a role to a user |
| `DELETE` | `/_management/permissions/user/{user}/role/{role}` | `rbac:RemoveUserRole` | Remove a role from a user |
| `PUT` | `/_management/permissions/user/{user}/role/{role}` | `rbac:ReplaceUserRoleScope` | Replace the scope of a user's role binding |

Request body (POST): `AuthorizationAPIUserRoleBindingCreateRequest { scope }` (optional scope constraint)
Request body (PUT): `AuthorizationAPIUserRoleBindingUpdateRequest { scope }` (new scope for the binding)
Response body (GET list): `AuthorizationAPIUserRoleBindingListResponse { bindings[] }` where each binding contains `{ role, scope }`
Response body (POST/PUT): `AuthorizationAPIUserRoleBindingResponse { user, role, scope }`
Response body (DELETE): empty (204 No Content)

Implementation: `pkg/sidecar/services/authorization/impl_api_user_role_bindings.go`

Bindings are keyed as `urb:<user>:<role>`. The list endpoint scans by prefix
`urb:<user>:` to find all bindings for a user.

## Integration API

### Authentication (`/_integration/authn/v1/`)

Exposed by the authentication integration service.

| Method | Path | Description |
|---|---|---|
| `GET` | `/identity` | Returns the current user's username |
| `POST` | `/validate` | Validates a token, returns user, groups, and lifetime |
| `POST` | `/login` | Authenticates with username/password, returns JWT |
| `GET` | `/logout` | Clears cookies, redirects to the login page (root `/`) |
| `POST` | `/createToken` | Creates a JWT for a specified user with groups |

**Request and response details:**

- **`/identity`** — No request body. Returns `IdentityResponse { user }` with the
  authenticated user's username extracted from the Bearer token or gRPC metadata.
- **`/validate`** — Request body: `ValidateRequest { token }` (the JWT string to
  validate). Returns `ValidateResponse { details { lifetime, user, groups[] } }`
  with the parsed token claims, or an error if the token is invalid/expired.
- **`/login`** — Request body: `LoginRequest { username, password }`. Returns
  `LoginResponse { token }` containing the signed JWT. Sets the token as an HTTP
  cookie for browser-based flows.
- **`/logout`** — No request body. Clears authentication cookies and returns an
  HTTP redirect (302) to the login page (root `/`). Returns no body.
- **`/createToken`** — Request body: `CreateTokenRequest { user, groups[], ttl }`.
  Returns `CreateTokenResponse { token }` containing a newly signed JWT for the
  specified user with the given groups and lifetime.

Proto: `integrations/authentication/v1/definition/definition.proto`
Implementation: `integrations/authentication/v1/implementation.go`

Key details:
- `Identity` extracts the user from the Bearer token or gRPC metadata
- `Validate` parses the JWT and returns `ValidateResponseDetails` with
  `lifetime`, `user`, and `groups[]`
- `CreateToken` signs a new JWT with the deployment's secret

### Authorization (`/_integration/authorization/v1/`)

Exposed by the authorization integration service. These are the programmatic
evaluation endpoints used by other services.

| Method | Path | Description |
|---|---|---|
| `POST` | `/evaluate` | Evaluate permission for user + groups + action/resource |
| `POST` | `/evaluate-many` | Batch evaluate multiple action/resource pairs |
| `POST` | `/evaluate-token` | Evaluate permission from a JWT token |
| `POST` | `/evaluate-token-many` | Batch evaluate from a JWT token |

Proto: `integrations/authorization/v1/definition/request.proto`
Implementation: `integrations/authorization/v1/impl.go`

**Request and response details:**

- **`/evaluate`** — Request body: `EvaluateRequest { user, groups[], action, resource, context }`.
  Returns `EvaluateResponse { effect }` (Allow or Deny).
- **`/evaluate-many`** — Request body: `EvaluateManyRequest { user, groups[], items[] }` where
  each item contains `{ action, resource, context }`. Returns `EvaluateManyResponse { results[] }`
  with an effect per item.
- **`/evaluate-token`** — Request body: `EvaluateTokenRequest { token, action, resource, context }`.
  Returns `EvaluateResponse { effect }`.
- **`/evaluate-token-many`** — Request body: `EvaluateTokenManyRequest { token, items[] }`.
  Returns `EvaluateManyResponse { results[] }`.

Key details:
- `Evaluate` / `EvaluateMany` accept explicit `user` + `groups` fields
- `EvaluateToken` / `EvaluateTokenMany` extract user and groups from the JWT
- Evaluation results are logged at the sidecar authorization layer (`sidecar-authz`):
  - **Deny** — logged at Info level with user, action, resource, and reason
  - **Allow** — logged at Debug level with user, action, resource
- In **central-permissive** mode, the Permissive wrapper adds its own logging
  (method=`Permissive.Evaluate`) before overriding denials to allows.
  The sanity check `POD.006` parses these logs to detect denied permissions.

### Mutation Logging and Auditing

Policy, role, and user-role-binding mutations are logged at Info level
(`sidecar-authz` logger) with the acting user and resource name:

- Policy created/updated/deleted
- Role created/updated/deleted
- User role assigned/removed/scope replaced
  using the token extractor (`pkg/util/token/extractor.go`)
- All use the configured plugin chain (superuser → central/permissive → policy eval)

#### Audit Log Format

Each mutation log entry includes:
- **Timestamp** — when the mutation occurred
- **Actor** — the authenticated user who performed the mutation (extracted from JWT)
- **Operation** — the type of mutation (create, update, delete, assign, remove)
- **Resource type** — policy, role, or user-role-binding
- **Resource name** — the name/key of the affected resource
- **Logger** — `sidecar-authz`

#### Accessing Audit Logs

Audit log entries are emitted to the sidecar container's standard output and can
be accessed via:
- `kubectl logs <pod> -c sidecar` — view live logs from the authorization sidecar
- Standard Kubernetes log aggregation pipelines (e.g., Fluentd, Loki, or
  Elasticsearch) that collect container logs
- Filter for logger `sidecar-authz` at Info level to isolate RBAC mutation events
- In **central-permissive** mode, additional evaluation audit entries are logged
  by the Permissive wrapper (method=`Permissive.Evaluate`) which can be used to
  review permission decisions that were overridden from deny to allow

## CRD Handlers

The binding CRDs have dedicated handler files (reconciliation not yet
implemented — CRD types and registration are in place):

| CRD | Type files | Handler |
|---|---|---|
| ArangoPermissionPolicy | `pkg/apis/permission/v1alpha1/policy*.go` | (via sidecar API) |
| ArangoPermissionRole | `pkg/apis/permission/v1alpha1/role*.go` | (via sidecar API) |
| ArangoPermissionToken | `pkg/apis/permission/v1alpha1/token*.go` | `pkg/handlers/permission/token/` |
| ArangoPermissionPolicyRoleBinding | `pkg/apis/permission/v1alpha1/policy_role_binding*.go` | `pkg/handlers/permission/policy_role_binding/` |
| ArangoPermissionRoleUserBinding | `pkg/apis/permission/v1alpha1/role_user_binding*.go` | (pending) |

Constants and resource names: `pkg/apis/permission/definitions.go`
Type registration: `pkg/apis/permission/v1alpha1/register.go`

## Pool Streaming (internal)

The `AuthorizationPoolService` provides server-streaming RPCs for client-side
cache synchronization. These are internal APIs used by the integration service,
not exposed externally.

| RPC | Description |
|-----|-------------|
| `GetPolicy` | Initial load of all policies |
| `PoolPolicyChanges` | Long-poll for policy changes |
| `GetRole` | Initial load of all groups (stored as roles) |
| `PoolRoleChanges` | Long-poll for group changes |
| `GetUserRoleBinding` | Initial load of all user-group bindings |
| `PoolUserRoleBindingChanges` | Long-poll for user-group binding changes |

Key files:
- `pkg/sidecar/services/authorization/definition/pool.proto`
- `pkg/sidecar/services/authorization/impl_client_role.go`
- `pkg/sidecar/services/authorization/impl_client_user_role_binding.go`
- `pkg/sidecar/services/authorization/client/client.go` — streaming client

## Proto Files

| Proto | Package | Service |
|---|---|---|
| `pkg/sidecar/services/authorization/definition/api.proto` | `service` | `AuthorizationAPI` |
| `pkg/sidecar/services/authorization/definition/pool.proto` | `service` | `AuthorizationPoolService` |
| `integrations/authentication/v1/definition/definition.proto` | `authentication` | `AuthenticationV1` |
| `integrations/authorization/v1/definition/definition.proto` | `authorization` | `AuthorizationV1` |
| `integrations/authorization/v1/definition/request.proto` | `authorization` | (messages only) |

## Proto Types

| Proto | Package |
|---|---|
| `pkg/sidecar/services/authorization/types/policy.proto` | `types` |
| `pkg/sidecar/services/authorization/types/role.proto` | `types` |
| `pkg/sidecar/services/authorization/types/user_role_binding.proto` | `types` |
| `pkg/sidecar/services/authorization/types/effect.proto` | `types` |
| `pkg/sidecar/services/authorization/types/context.proto` | `types` |
