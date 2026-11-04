---
layout: page
title: Policies and Roles
parent: RBAC
grand_parent: ArangoDBPlatform
nav_order: 3
---

# Policies and Roles

## Policies

A policy defines what actions are allowed or denied on which resources.

### Structure

```yaml
statements:
  - effect: Allow        # or Deny
    actions:
      - "db:Read"
      - "db:WriteData"
    resources:
      - "db:database:sales"
      - "db:collection:sales:*"
```

### Actions

Actions use the format `<namespace>:<name>`, or `*` for all actions. Each
component defines the actions it checks under its own namespace:

| Namespace | Checked by | Examples |
|---|---|---|
| `db` | ArangoDB (CoreDB) | `db:Read`, `db:WriteData`, `db:WriteMeta`, `db:Create`, `db:Drop`, `db:UseApiVersion`, `db:Admin*` |
| `rbac` | Authorization sidecar (management API) | `rbac:ListRole`, `rbac:AssignUserRole` |
| `ai` | AI services / Arango Control Plane | `ai:acp.workspace.read`, `ai:acp.secret.manage` |

The main ArangoDB actions:

| Action | Allows |
|---|---|
| `db:UseApiVersion` | Using the REST API at all. **Every** authenticated request to ArangoDB checks it on `db:apiversion:<version>` |
| `db:Read` | Seeing and reading databases, collections, views, analyzers, graphs, users |
| `db:WriteData` | Writing documents to collections |
| `db:WriteMeta` | Changing properties/metadata (e.g. indexes, view or graph definitions) |
| `db:Create`, `db:Drop` | Creating and dropping databases, collections, views, analyzers, graphs, users |
| `db:Admin*` | Server administration (`db:AdminBackup`, `db:AdminReadLogs`, `db:AdminReadUsers`, ...) |

### Resources

ArangoDB resources are colon-separated paths:

| Resource | Example |
|---|---|
| `db:database:<db>` | `db:database:sales` |
| `db:collection:<db>:<collection>` | `db:collection:sales:orders` |
| `db:view:<db>:<view>` | `db:view:sales:ordersView` |
| `db:analyzer:<db>:<analyzer>` | `db:analyzer:sales:text_en` |
| `db:graph:<db>:<graph>` | `db:graph:sales:customers` |
| `db:user:<user>` | `db:user:alice` |
| `db:apiversion:<version>` | `db:apiversion:1` |

### Pattern Matching

Patterns are matched segment by segment, split at `:`:

- A segment is `*` (any value), `<prefix>*`, `*<suffix>`, or an exact value.
- The pattern must have the same number of segments as the value, unless the
  pattern's last segment is `*`, which also matches any further segments.
- An action pattern must be `*` or have exactly two segments.

| Pattern | Matches | Does not match |
|---|---|---|
| `*` | everything | |
| `db:*` | every `db` action or resource | `ai:...` |
| `db:Admin*` | `db:AdminBackup`, `db:AdminReadLogs` | `db:Read` |
| `*:Read` | `db:Read` | `db:WriteData` |
| `db:database:sales` | `db:database:sales` | `db:database:sales2` |
| `db:collection:sales:*` | all collections in `sales` | `db:collection:hr:staff` |
| `db:*:sales:*` | all collections, views, analyzers and graphs in `sales` | `db:database:sales` (fewer segments) |
| `db:collection:sales:tmp_*` | `db:collection:sales:tmp_2024` | `db:collection:sales:orders` |

### Evaluation Rules

- **Deny-by-default** - If no statement matches, the request is denied
- **Explicit deny wins** - A Deny statement overrides any Allow
- **Order does not matter** - All statements are evaluated; first Deny found
  stops evaluation

> **Note:** Because ArangoDB checks `db:UseApiVersion` on every request, every
> user who should access ArangoDB needs it, both in the role's policies and in
> the scope of their binding. A binding scope that only lists databases and
> collections denies all requests. Include `db:apiversion:*` (or `*`) in the
> scope's resources.

## Roles

A role is a named container that groups one or more policies. A role does not
define a scope of its own - the scope boundary is set per user-role assignment
(see `ArangoPermissionRoleUserBinding` and tokens below).

### Structure

A role contains only a `deployment` reference and an optional `description`.

Roles do not directly reference named policies. To attach a policy to a role,
use an `ArangoPermissionPolicyRoleBinding` (see below).

## Sidecar Resource Naming

When Kubernetes CRDs are reconciled into the sidecar authorization service,
policies and roles are stored with `managed:operator:` prefixed names:

| CRD | Sidecar Name | Example | Notes |
|---|---|---|---|
| ArangoPermissionPolicy | `managed:operator:<crd-name>` | `managed:operator:read-only` | |
| ArangoPermissionRole | `managed:operator:<crd-uid>` | `managed:operator:a1b2c3d4-...` | |
| ArangoPermissionToken (role) | `managed:operator:<crd-uid>` | `managed:operator:e5f6g7h8-...` | Created only when `spec.policy` is set; references the token's managed policy |
| ArangoPermissionToken (policy) | `managed:operator:<crd-uid>:policy` | `managed:operator:e5f6g7h8-...:policy` | Created from the inline `spec.policy` |
| Predefined roles | `managed:predefined:<id>` | `managed:predefined:coredb-reader` | Seeded by the operator, see [Predefined Roles](predefined_roles.md) |

### How Policies and Roles are Combined

When the Role handler reconciles an `ArangoPermissionRole`, it assembles the
sidecar role's `policies` list from **PolicyRoleBindings** — the handler lists
all `ArangoPermissionPolicyRoleBinding` CRDs that reference this role. For each
ready binding, the bound `ArangoPermissionPolicy`'s sidecar name
(`managed:operator:<policy-crd-name>`) is added to the list.

For example, given:
- `ArangoPermissionPolicy` named `read-only` (sidecar: `managed:operator:read-only`)
- `ArangoPermissionPolicy` named `write-reports` (sidecar: `managed:operator:write-reports`)
- `ArangoPermissionRole` named `editor`
- Two `ArangoPermissionPolicyRoleBinding` CRDs binding both policies to `editor`

The resulting sidecar role `managed:operator:<editor-uid>` will have:
```json
{
  "policies": [
    "managed:operator:read-only",
    "managed:operator:write-reports"
  ]
}
```

During evaluation, the role's policies are resolved and their statements are
evaluated together, bounded by the scope of the user-role binding (or token)
that grants the role. An explicit Deny in any policy overrides Allows from others.

## Binding CRDs

Two binding CRDs connect policies, roles, and users together:

- **ArangoPermissionPolicyRoleBinding** - Binds a named policy to a role
- **ArangoPermissionRoleUserBinding** - Binds a role to a user with a scope

### Full Flow

The recommended flow for setting up RBAC via CRDs is:

1. Create a policy (`ArangoPermissionPolicy`)
2. Create a role (`ArangoPermissionRole`)
3. Bind the policy to the role (`ArangoPermissionPolicyRoleBinding`)
4. Bind the role to a user (`ArangoPermissionRoleUserBinding`)

## Managing via API

> **Access Control:** All management API endpoints require authentication and
> the appropriate RBAC permission (e.g. `rbac:CreatePolicy` to create a policy).
> In the initial release, these endpoints are intended for **administrator use
> only**. Regular users should not be granted `rbac:*` permissions. Use CRDs
> (below) for declarative management in production.

### Create a policy

```bash
curl -X POST https://<gateway>/_management/permissions/policy/read-only \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "item": {
      "statements": [
        {
          "effect": "Allow",
          "actions": ["db:UseApiVersion"],
          "resources": ["*"]
        },
        {
          "effect": "Allow",
          "actions": ["db:Read"],
          "resources": ["db:database:*", "db:collection:*:*"]
        }
      ]
    }
  }'
```

### Create a role

```bash
curl -X POST https://<gateway>/_management/permissions/role/viewer \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "item": {
      "policies": ["read-only"],
      "description": "Read-only access"
    }
  }'
```

A role only references policies. Its `scope` field is deprecated and ignored
during evaluation. Set the scope on the
[user role binding](user_bindings.md) or [group role binding](group_bindings.md)
instead.

### List policies

```bash
curl -H "Authorization: Bearer $TOKEN" \
  https://<gateway>/_management/permissions/policy
```

### List roles

```bash
curl -H "Authorization: Bearer $TOKEN" \
  https://<gateway>/_management/permissions/role
```

List endpoints return names only, e.g.
`{"names": ["read-only"], "page": {"items": 10, "page": 0}}`. They accept the
optional query parameters `items_per_page` and `page`. Use
`GET .../policy/{name}` or `GET .../role/{name}` to fetch the details.

### Endpoint reference

| Method | Path | RBAC action |
|---|---|---|
| GET | `/_management/permissions/policy` | `rbac:ListPolicy` |
| GET | `/_management/permissions/policy/{name}` | `rbac:GetPolicy` |
| POST | `/_management/permissions/policy/{name}` | `rbac:CreatePolicy` |
| PUT | `/_management/permissions/policy/{name}` | `rbac:UpdatePolicy` |
| DELETE | `/_management/permissions/policy/{name}` | `rbac:DeletePolicy` |
| GET | `/_management/permissions/role` | `rbac:ListRole` |
| GET | `/_management/permissions/role/{name}` | `rbac:GetRole` |
| POST | `/_management/permissions/role/{name}` | `rbac:CreateRole` |
| PUT | `/_management/permissions/role/{name}` | `rbac:UpdateRole` |
| DELETE | `/_management/permissions/role/{name}` | `rbac:DeleteRole` |
| POST | `/_management/permissions/refresh` | none (any authenticated caller) |

`POST /_management/permissions/refresh` makes the receiving sidecar re-read
all policies, roles and bindings from the database and returns once that is
done. Use it to skip the regular refresh interval.

Objects created through the API keep the exact name you use. Objects
reconciled from CRDs get a `managed:operator:` prefix (see
[Sidecar Resource Naming](#sidecar-resource-naming)). Don't change
`managed:*` objects through the API; change the CRDs instead, because the
operator manages these objects.

## Managing via CRDs

Policies, roles, and bindings can be managed through Kubernetes custom resources.

### Step 1: Create a Policy (ArangoPermissionPolicy)

```yaml
apiVersion: permission.arangodb.com/v1alpha1
kind: ArangoPermissionPolicy
metadata:
  name: read-only
spec:
  deployment:
    name: my-deployment
  policy:
    statements:
      - effect: Allow
        actions:
          - "db:UseApiVersion"
        resources:
          - "*"
      - effect: Allow
        actions:
          - "db:Read"
        resources:
          - "db:database:*"
          - "db:collection:*:*"
```

### Step 2: Create a Role (ArangoPermissionRole)

```yaml
apiVersion: permission.arangodb.com/v1alpha1
kind: ArangoPermissionRole
metadata:
  name: viewer
spec:
  deployment:
    name: my-deployment
```

### Step 3: Bind the Policy to the Role (ArangoPermissionPolicyRoleBinding)

An `ArangoPermissionPolicyRoleBinding` attaches a named policy to a role.
This is how you compose roles from reusable policies.

```yaml
apiVersion: permission.arangodb.com/v1alpha1
kind: ArangoPermissionPolicyRoleBinding
metadata:
  name: viewer-read-only-binding
spec:
  deployment:
    name: my-deployment
  policy:
    name: read-only
  role:
    name: viewer
```

| Field | Required | Description |
|---|---|---|
| `spec.deployment` | Yes | Reference to the ArangoDB deployment |
| `spec.policy.name` | Yes | Name of the ArangoPermissionPolicy CRD (resolved to sidecar name) |
| `spec.role.name` | Yes | Name of the ArangoPermissionRole CRD (resolved to sidecar name) |

A role can have multiple policy bindings. Each binding adds the referenced
policy to the role's evaluation set.

### Step 4: Bind the Role to a User (ArangoPermissionRoleUserBinding)

An `ArangoPermissionRoleUserBinding` assigns a role to a specific user with
an inline scope policy. The scope restricts what the role grants for this
particular user. It can never grant more than the role's policies.

```yaml
apiVersion: permission.arangodb.com/v1alpha1
kind: ArangoPermissionRoleUserBinding
metadata:
  name: alice-viewer-binding
spec:
  deployment:
    name: my-deployment
  role:
    name: viewer
  userName: alice
  scope:
    statements:
      - effect: Allow
        actions:
          - "*"
        resources:
          - "db:apiversion:*"
          - "db:database:reports"
          - "db:collection:reports:*"
```

| Field | Required | Description |
|---|---|---|
| `spec.deployment` | Yes | Reference to the ArangoDB deployment |
| `spec.role.name` | Yes | Name of the ArangoPermissionRole CRD (resolved to sidecar name) |
| `spec.userName` | Yes | Username to assign the role to |
| `spec.scope` | Yes | Inline policy scoping this user-role assignment |

### Complete Example

The following set of resources grants user `alice` read-only access to the
`reports` database and its collections:

```yaml
---
apiVersion: permission.arangodb.com/v1alpha1
kind: ArangoPermissionPolicy
metadata:
  name: read-only
spec:
  deployment:
    name: my-deployment
  policy:
    statements:
      - effect: Allow
        actions:
          - "db:UseApiVersion"
        resources:
          - "*"
      - effect: Allow
        actions:
          - "db:Read"
        resources:
          - "db:database:*"
          - "db:collection:*:*"
---
apiVersion: permission.arangodb.com/v1alpha1
kind: ArangoPermissionRole
metadata:
  name: viewer
spec:
  deployment:
    name: my-deployment
---
apiVersion: permission.arangodb.com/v1alpha1
kind: ArangoPermissionPolicyRoleBinding
metadata:
  name: viewer-read-only-binding
spec:
  deployment:
    name: my-deployment
  policy:
    name: read-only     # references ArangoPermissionPolicy CRD
  role:
    name: viewer        # references ArangoPermissionRole CRD
---
apiVersion: permission.arangodb.com/v1alpha1
kind: ArangoPermissionRoleUserBinding
metadata:
  name: alice-viewer-binding
spec:
  deployment:
    name: my-deployment
  role:
    name: viewer
  userName: alice
  scope:
    statements:
      - effect: Allow
        actions:
          - "*"
        resources:
          - "db:apiversion:*"
          - "db:database:reports"
          - "db:collection:reports:*"
```
