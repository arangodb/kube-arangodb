# Predefined Roles

> **Alpha Feature** - RBAC is currently in alpha (`v1alpha1`). The predefined role
> catalog, its names, and the accesses it grants may change without notice.

The operator seeds every RBAC-enabled deployment with a catalog of **predefined
roles**. These roles are operator-managed: customers can assign and scope them,
but cannot edit or delete them. Only the roles listed here are shipped - users
cannot define their own predefined roles (custom roles are still possible as
regular `ArangoPermissionRole` objects, see
[docs/platform/rbac/predefined_roles.md](../../docs/platform/rbac/predefined_roles.md)).

## Model

- **Default deny.** No access is granted unless a scoped role explicitly allows
  it. A predefined role grants nothing until it is bound to a user (via an
  `ArangoPermissionRoleUserBinding`) with a scope.
- **Predefined only (MVP).** Users assign and scope roles; they do not create
  predefined roles.
- **Deny and allow effects** with deny taking precedence.
- **Scoping** is applied per user-role binding at resource-type level, with one,
  many, and pattern (URN) matching (see [Data Model](data_model.md)).

### Granularity of resource types

| Component | Resource types |
|---|---|
| CoreDB | `Database`, `Collection` |
| AI Services | `Workspace` (equivalent to `Database`) |

## Naming

Predefined roles are created in the authorization sidecar under the reserved
prefix `managed:predefined:`, following the `managed:` convention used for other
operator-owned objects (handlers use `managed:operator:`). A role and its bundled
policy share the same name (they live in separate policy/role collections):

```
managed:predefined:<role>
```

For example, `managed:predefined:coredb-reader`. The prefix marks the object as
operator-managed, which is why it is visible but not editable.

## Catalog

| Role (`managed:predefined:…`) | Description | Access |
|---|---|---|
| `super-admin` | Reserved role providing full access to all functionality. **Cannot be assigned by the customer.** | Allow `*` on `*`. Bound automatically to the deployment `root` user with an Allow-all scope. |
| `tenant-admin` | Manages users and role bindings. | Assign and scope roles; read users, roles, and resources. |
| `coredb-reader` | Reads scoped resources and executes read-only database operations. | Read-only on scoped `Database`/`Collection`. |
| `coredb-developer` | Reads and writes scoped resources and executes read and write database operations. | Read and write on scoped `Database`/`Collection`. |
| `coredb-admin` | Manages scoped resources' structures and lifecycle. | Create, alter, and drop scoped `Database`/`Collection`. |
| `ai-user` | Executes AI workflows and reads resulting outputs within scoped resources. | Execute AI workflows and read outputs on scoped `Workspace`. |
| `ai-developer` | Builds, configures, manages, and executes AI workflows and artifacts within scoped resources. | Full AI workflow and artifact management on scoped `Workspace`. |
| `platform-operator` | Operates platform services, manages bundled services, views observability, and starts containers within scoped resources. | Operate platform/bundled services, view observability, start containers on scoped resources. |
| `secret-admin` | Manages secrets within scoped resources. | Manage secrets on scoped resources. |

### Bundled policies (MVP status)

The **Access** column above describes each role's intended authorization. In the
current MVP only `super-admin` ships a concrete bundled policy (Allow `*` on `*`).
The remaining roles are created as **empty role containers** - they exist so they
can be assigned, scoped, and extended, and so their names are stable across
releases, but their bundled policies are defined in a later iteration. Until a
bundled policy is defined (or a policy is attached to the role, see
[Extending](#extending)), binding one of these roles grants no permissions
(default deny).

Bundled third-party services and solutions are exposed as **binary (on/off)**
permissions; the concrete action set for each service is service-defined.

## Lifecycle

Predefined roles are synced by the deployment reconciler:

1. After the deployment bootstrap completes (the `root` user is created) and the
   gateway/authorization sidecar is enabled, the reconciler runs the
   `SyncRBACPermissions` action.
2. The action connects to the authorization sidecar and upserts every predefined
   role directly through the management API - it does **not** create
   intermediate `ArangoPermission*` custom resources. Each role's policy set is
   its bundled policy (for `super-admin`) merged with any policies attached via
   `ArangoPermissionPolicyRoleBindings` (see [Extending](#extending)).
3. `super-admin` additionally gets its Allow-all policy and a binding of the
   `root` user with an Allow-all scope.
4. A throttled high plan builder re-runs the action periodically, so a deleted or
   drifted predefined role is recreated and repaired.

See `pkg/deployment/reconcile/plan_builder_rbac.go` for the catalog and
`plan_builder_high.go` (`createSyncRBACPermissionsPlan`) for the builder.

## Extending

Predefined roles cannot be renamed or deleted, but they are **extensible**:

- **Attach policies to a predefined role.** An `ArangoPermissionPolicyRoleBinding`
  may reference a predefined role through the reference's `direct` field, which
  holds the exact sidecar name (e.g. `managed:predefined:coredb-reader`). There is
  no `ArangoPermissionRole` CRD, so `direct` is used instead of `name`; the
  `role_user_binding`/`policy_role_binding` handlers resolve it as-is
  (`ArangoPermissionBindingRef.IsDirect`). The reconciler discovers such bindings
  (`collectBoundPolicies`), resolves their policies to sidecar names, and merges
  them into the predefined role. Drift-repair targets the merged set, so bound
  policies are preserved; unbinding removes them again.
- **Per-user composition.** A user's effective permissions are the union of all
  their role bindings (with `Deny` precedence), so a user can be granted an
  additional custom `ArangoPermissionRole` on top of a predefined one.

See [docs/platform/rbac/predefined_roles.md](../../docs/platform/rbac/predefined_roles.md#extending-predefined-roles).

## Machine users

Services that must operate independently of a human user's permissions
authenticate as machine identities. The operator itself acts as a machine
identity (`managed:operator:*`) when reconciling authorization state, and
services obtain scoped bearer tokens through `ArangoPermissionToken`
(see [Token Integration](token_integration.md)).
