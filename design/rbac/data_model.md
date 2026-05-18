# RBAC Data Model

## Policy

A Policy contains a list of statements. Each statement has an effect
(Allow/Deny), a list of actions, and a list of resources.

```protobuf
message Policy {
  repeated PolicyStatement statements = 1;
}

message PolicyStatement {
  Effect effect = 1;
  repeated string resources = 2;
  repeated string actions = 3;
}

enum Effect {
  Deny = 0;
  Allow = 1;
}
```

### Action Format

Actions use the format `<namespace>:<name>` (e.g. `rbac:ListRole`,
`database:write`). Supported patterns:

| Pattern | Example | Matches |
|---|---|---|
| Exact | `rbac:ListRole` | Only `rbac:ListRole` |
| Wildcard all | `*` | Everything |
| Prefix | `rbac:*` | `rbac:ListRole`, `rbac:CreatePolicy`, etc. |
| Suffix | `*:List*` | `rbac:ListRole`, `rbac:ListPolicy`, etc. |

Resources use the same wildcard matching.

### Validation

- Actions must be `*` or `<namespace>:<name>` format
- Resources must be non-empty strings
- Effect must be `Allow` or `Deny`

Key files:
- `pkg/sidecar/services/authorization/types/policy.proto`
- `pkg/sidecar/services/authorization/types/policy.go`
- `pkg/sidecar/services/authorization/client/policy.go`

## Role

A Role is a named collection of policy references, managed via CRDs. Roles are
the building block — they do not carry scope. Scope is applied when a role is
assigned to a group (via `ArangoPermissionScopedBindingRef` on the token).

Key files:
- `pkg/apis/permission/v1alpha1/role_spec.go`

## Group (sidecar Role with Scope)

A Group is the evaluation unit carried in JWTs. In the sidecar storage it is
stored as a `Role` proto with policies + scope. The scope is required — groups
without a scope are skipped during evaluation.

```protobuf
message Role {
  repeated string policies = 1;
  Policy scope = 3;
}
```

- `policies` — List of named policy references (resolved from the `_policies` collection)
- `scope` — Required inline boundary Policy. The scope acts as an intersection
  filter: a permission is only granted when **both** the named policies allow it
  **and** the scope allows it. Groups without a scope are skipped during
  evaluation and do not contribute to the result

Key files:
- `pkg/sidecar/services/authorization/types/role.proto`
- `pkg/sidecar/services/authorization/types/role.go`

## UserRoleBinding

A UserRoleBinding assigns a role to a specific user with an inline scope
policy. This allows the same role to be scoped differently per user.

```protobuf
message UserRoleBinding {
  string role = 1;
  Policy scope = 2;
}
```

Both `role` and `scope` are required fields. Bindings are stored in the
`_user_role_bindings` ArangoDB collection, keyed as `urb:<user>:<role>`.

Key files:
- `pkg/sidecar/services/authorization/types/user_role_binding.proto` — Proto definition
- `pkg/sidecar/services/authorization/types/user_role_binding.pb.go` — Generated Go type
- `pkg/sidecar/services/authorization/types/user_role_binding.go` — Hash, Clean, Validate

## Sidecar Resource Naming Convention

When CRDs are reconciled into the sidecar, resources are stored with a
`managed:operator:` prefix to distinguish operator-managed resources from
resources created directly via the API:

| CRD | Sidecar Name Pattern |
|---|---|
| ArangoPermissionPolicy | `managed:operator:<crd-name>` |
| ArangoPermissionRole | `managed:operator:<crd-uid>` |
| ArangoPermissionToken (role) | `managed:operator:<crd-uid>` |

Policies use the CRD **name** (human-readable, must be unique in namespace).
Roles use the CRD **UID** (guaranteed unique, avoids collisions on recreate).

### Role Policies List Assembly

The Role handler builds the sidecar role's `policies` list entirely from
`ArangoPermissionPolicyRoleBinding` CRDs. For each ready binding targeting
this role, the bound policy's sidecar name (`managed:operator:<policy-crd-name>`)
is added to the list.

This means modifying bindings (create/delete) causes the Role handler to
re-render the sidecar role with the updated policies list on next reconcile.

## CRD to Proto Mapping

The following Kubernetes CRDs map to the internal proto types:

| CRD | Proto Type | Description |
|---|---|---|
| ArangoPermissionPolicy | Policy | Defines a named policy with statements |
| ArangoPermissionRole | _(operational)_ | Defines a role (policy container); policies attached via PolicyRoleBinding |
| ArangoPermissionPolicyRoleBinding | _(operational)_ | Binds a Policy to a Role by name; adds the policy name to the Role's `policies` list |
| ArangoPermissionRoleUserBinding | UserRoleBinding | Binds a Role to a user with an inline scope; creates a UserRoleBinding in the sidecar |
| ArangoPermissionToken | Role (as group) | References roles via `spec.roles[]` with scoped bindings; auto-creates managed groups in the sidecar |

### ArangoPermissionPolicyRoleBinding

This CRD does not map directly to a single proto type. Instead, the Role
handler reads all ready bindings targeting a role and assembles the full
`policies` list on the sidecar role document. The CRD spec contains:

- `deployment` (required) - Reference to the ArangoDB deployment
- `policy` (required) - Object reference to the ArangoPermissionPolicy
- `role` (required) - Object reference to the ArangoPermissionRole

The binding handler validates that both the referenced Policy and Role exist
and are ready. On deletion, the finalizer ensures the binding is cleanly
removed, triggering the Role handler to rebuild its policies list without
the removed binding.

Key files:
- `pkg/apis/permission/v1alpha1/policy_role_binding.go` — CRD type definition
- `pkg/apis/permission/v1alpha1/policy_role_binding_spec.go` — Spec with validation
- `pkg/apis/permission/v1alpha1/policy_role_binding_status.go` — Status
- `pkg/handlers/permission/policy_role_binding/handler.go` — Reconciliation handler
- `pkg/handlers/permission/policy_role_binding/handler_references.go` — Policy/Role resolution

### ArangoPermissionRoleUserBinding

This CRD maps to the `UserRoleBinding` proto type. Its reconciliation creates
a UserRoleBinding entry in the sidecar's `_user_role_bindings` collection.
The CRD spec contains:

- `deployment` (required) - Reference to the ArangoDB deployment
- `role` (required) - Role reference (`name` for CRD name)
- `userName` (required) - Username to assign the role to
- `scope` (required) - Inline policy (statements with effect/actions/resources)

Key files:
- `pkg/apis/permission/v1alpha1/role_user_binding.go` — CRD type definition
- `pkg/apis/permission/v1alpha1/role_user_binding_spec.go` — Spec with validation
- `pkg/apis/permission/v1alpha1/role_user_binding_status.go` — Status

## Relationships

```
ArangoPermissionRoleUserBinding ----> ArangoPermissionRole ----> ArangoPermissionPolicy (via PolicyRoleBinding)
         |                                     |
         |                                     +------> Policy (inline scope on Role)
         |
         +-----> Policy (inline scope on binding)

CRD linkage:

  ArangoPermissionPolicy
         ^
         | (policy.name)
  ArangoPermissionPolicyRoleBinding
         | (role.name)
         v
  ArangoPermissionRole
         ^
         | (role.name)
  ArangoPermissionRoleUserBinding
         | (userName)
         v
       User
```

When evaluating permissions for a user, the system resolves group names from the
JWT and builds a `ScopedPolicy` per group:
1. Named policies referenced by the group (from the sidecar role's `policies` list)
2. The group's scope as the boundary
3. A permission is granted only when the policies allow it AND the scope allows it
4. Groups without a scope are skipped entirely
