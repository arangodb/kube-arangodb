---
layout: page
title: Policies and Roles
parent: RBAC
grand_parent: ArangoDBPlatform
nav_order: 2
---

# Policies and Roles

## Policies

A policy defines what actions are allowed or denied on which resources.

### Structure

```yaml
statements:
  - effect: Allow        # or Deny
    actions:
      - "database:read"
      - "collection:*"
    resources:
      - "mydb"
      - "reports-*"
```

### Actions

Actions use the format `<namespace>:<name>`:

| Example | Description |
|---|---|
| `database:read` | Read access to databases |
| `collection:write` | Write access to collections |
| `rbac:ListRole` | List authorization roles |
| `*` | All actions |
| `database:*` | All database actions |

### Resources

Resources identify what the action applies to. They support the same wildcard
patterns as actions:

| Pattern | Matches |
|---|---|
| `mydb` | Exactly `mydb` |
| `*` | Everything |
| `reports-*` | `reports-daily`, `reports-weekly`, etc. |

### Evaluation Rules

- **Deny-by-default** - If no statement matches, the request is denied
- **Explicit deny wins** - A Deny statement overrides any Allow
- **Order does not matter** - All statements are evaluated; first Deny found
  stops evaluation

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
| ArangoPermissionToken (role) | `managed:operator:<crd-uid>` | `managed:operator:e5f6g7h8-...` | References named policies via `spec.policies[]` and an optional inline `spec.policy` |

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
          "actions": ["database:read", "collection:read"],
          "resources": ["*"]
        }
      ]
    }
  }'
```

### Create a role

```bash
curl -X POST https://<gateway>/_management/permissions/role/viewer \
  -d '{
    "item": {
      "policies": ["read-only"],
      "scope": {
        "statements": [
          {
            "effect": "Allow",
            "actions": ["database:read"],
            "resources": ["*"]
          }
        ]
      }
    }
  }'
```

### List policies

```bash
curl https://<gateway>/_management/permissions/policy
```

### List roles

```bash
curl https://<gateway>/_management/permissions/role
```

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
          - "database:read"
          - "collection:read"
        resources:
          - "*"
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
an inline scope policy. The scope restricts or extends what the role grants
for this particular user.

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
          - "database:read"
          - "collection:read"
        resources:
          - "reports"
```

| Field | Required | Description |
|---|---|---|
| `spec.deployment` | Yes | Reference to the ArangoDB deployment |
| `spec.role.name` | Yes | Name of the ArangoPermissionRole CRD (resolved to sidecar name) |
| `spec.userName` | Yes | Username to assign the role to |
| `spec.scope` | Yes | Inline policy scoping this user-role assignment |

### Complete Example

The following set of resources grants user `alice` read-only access scoped to
the `reports` resource:

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
          - "database:read"
          - "collection:read"
        resources:
          - "*"
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
          - "database:read"
          - "collection:read"
        resources:
          - "reports"
```
