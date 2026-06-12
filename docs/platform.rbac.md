---
layout: page
has_children: true
title: RBAC
parent: ArangoDBPlatform
nav_order: 6
---

# Platform RBAC

Role-Based Access Control (RBAC) provides policy-based authorization for
requests to ArangoDB deployments managed by the platform. When enabled,
every request through the gateway is evaluated against policies and roles
before reaching the database.

## CRD Overview

| CRD | Purpose |
|---|---|
| ArangoPermissionPolicy | Defines a reusable policy (statements with effect/actions/resources) |
| ArangoPermissionRole | Defines a role with an inline scope policy |
| ArangoPermissionPolicyRoleBinding | Binds a named policy to a role |
| ArangoPermissionRoleUserBinding | Binds a role to a user with a per-user scope |
| ArangoPermissionToken | Creates JWT tokens referencing an ArangoPermissionPolicy with an inline scope |

## Sections

- [Enabling RBAC](platform/rbac/enabling.md) - Feature flags, Helm configuration, authorization modes
- [Policies and Roles](platform/rbac/policies.md) - Defining permissions with policies, roles, bindings, and scopes
- [Permission Tokens](platform/rbac/tokens.md) - Creating JWT tokens via ArangoPermissionToken CRD
- [User Role Bindings](platform/rbac/user_bindings.md) - Assigning roles to users with per-user scopes
- [Identity and Permissions](platform/rbac/identity.md) - Who Am I, Can I, and authentication endpoints
- [FAQ](platform/rbac/faq.md) - Common questions and troubleshooting
