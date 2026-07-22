---
layout: page
parent: CRD reference
title: ArangoPermissionRoleUserBinding V1Alpha1
---

# API Reference for ArangoPermissionRoleUserBinding V1Alpha1

## Spec

### .spec.deployment.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.4/pkg/apis/shared/v1/object.go#L53)</sup>

This field is **required**

Name of the object

***

### .spec.role.direct

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.4/pkg/apis/permission/v1alpha1/binding_ref.go#L40)</sup>

Direct references an existing authorization object (role or policy) by its exact name, without
a backing ArangoPermission CRD - e.g. an operator-managed predefined role
"managed:predefined:coredb-reader". The value is used as-is. Exactly one of Name or Direct
must be set.

***

### .spec.role.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.4/pkg/apis/permission/v1alpha1/binding_ref.go#L34)</sup>

Name references an ArangoPermission CRD by name. The operator resolves it to the sidecar name.

***

### .spec.scope.description

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.4/pkg/apis/permission/v1alpha1/policy/policy.go#L31)</sup>

Description is an optional human-readable description of this policy

***

### .spec.scope.statements\[int\].actions

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.4/pkg/apis/permission/v1alpha1/policy/statement.go#L52)</sup>

This field is **required**

Actions defines the list of actions.
Action needs to be defined in format `<namespace>:<name>`

***

### .spec.scope.statements\[int\].description

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.4/pkg/apis/permission/v1alpha1/policy/statement.go#L41)</sup>

Description is an optional human-readable description of what this statement does

***

### .spec.scope.statements\[int\].effect

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.4/pkg/apis/permission/v1alpha1/policy/statement.go#L47)</sup>

This field is **required**

Effect defines the statement effect.

Possible Values: 
* `"Allow"` (default) - Action is Allowed
* `"Deny"` - Action is Denied

***

### .spec.scope.statements\[int\].resources

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.4/pkg/apis/permission/v1alpha1/policy/statement.go#L56)</sup>

This field is **required**

Resources defines the list of resources

***

### .spec.userName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.4/pkg/apis/permission/v1alpha1/role_user_binding_spec.go#L45)</sup>

This field is **required**

UserName is the name of the user to bind the role to

