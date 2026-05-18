---
layout: page
parent: CRD reference
title: ArangoPermissionToken V1Alpha1
---

# API Reference for ArangoPermissionToken V1Alpha1

## Spec

### .spec.deployment.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.3/pkg/apis/shared/v1/object.go#L53)</sup>

This field is **required**

Name of the object

***

### .spec.policy.description

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.3/pkg/apis/permission/v1alpha1/policy/policy.go#L31)</sup>

Description is an optional human-readable description of this policy

***

### .spec.policy.statements\[int\].actions

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.3/pkg/apis/permission/v1alpha1/policy/statement.go#L52)</sup>

This field is **required**

Actions defines the list of actions.
Action needs to be defined in format `<namespace>:<name>`

***

### .spec.policy.statements\[int\].description

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.3/pkg/apis/permission/v1alpha1/policy/statement.go#L41)</sup>

Description is an optional human-readable description of what this statement does

***

### .spec.policy.statements\[int\].effect

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.3/pkg/apis/permission/v1alpha1/policy/statement.go#L47)</sup>

This field is **required**

Effect defines the statement effect.

Possible Values: 
* `"Allow"` (default) - Action is Allowed
* `"Deny"` - Action is Denied

***

### .spec.policy.statements\[int\].resources

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.3/pkg/apis/permission/v1alpha1/policy/statement.go#L56)</sup>

This field is **required**

Resources defines the list of resources

***

### .spec.roles\[int\].role.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.3/pkg/apis/permission/v1alpha1/binding_ref.go#L33)</sup>

Name references an ArangoPermission CRD by name. The operator resolves it to the sidecar name.

***

### .spec.roles\[int\].scope.policy.description

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.3/pkg/apis/permission/v1alpha1/policy/policy.go#L31)</sup>

Description is an optional human-readable description of this policy

***

### .spec.roles\[int\].scope.policy.statements\[int\].actions

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.3/pkg/apis/permission/v1alpha1/policy/statement.go#L52)</sup>

This field is **required**

Actions defines the list of actions.
Action needs to be defined in format `<namespace>:<name>`

***

### .spec.roles\[int\].scope.policy.statements\[int\].description

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.3/pkg/apis/permission/v1alpha1/policy/statement.go#L41)</sup>

Description is an optional human-readable description of what this statement does

***

### .spec.roles\[int\].scope.policy.statements\[int\].effect

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.3/pkg/apis/permission/v1alpha1/policy/statement.go#L47)</sup>

This field is **required**

Effect defines the statement effect.

Possible Values: 
* `"Allow"` (default) - Action is Allowed
* `"Deny"` - Action is Denied

***

### .spec.roles\[int\].scope.policy.statements\[int\].resources

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.3/pkg/apis/permission/v1alpha1/policy/statement.go#L56)</sup>

This field is **required**

Resources defines the list of resources

***

### .spec.roles\[int\].scope.ref.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.3/pkg/apis/permission/v1alpha1/binding_ref.go#L33)</sup>

Name references an ArangoPermission CRD by name. The operator resolves it to the sidecar name.

***

### .spec.scope.description

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.3/pkg/apis/permission/v1alpha1/policy/policy.go#L31)</sup>

Description is an optional human-readable description of this policy

***

### .spec.scope.statements\[int\].actions

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.3/pkg/apis/permission/v1alpha1/policy/statement.go#L52)</sup>

This field is **required**

Actions defines the list of actions.
Action needs to be defined in format `<namespace>:<name>`

***

### .spec.scope.statements\[int\].description

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.3/pkg/apis/permission/v1alpha1/policy/statement.go#L41)</sup>

Description is an optional human-readable description of what this statement does

***

### .spec.scope.statements\[int\].effect

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.3/pkg/apis/permission/v1alpha1/policy/statement.go#L47)</sup>

This field is **required**

Effect defines the statement effect.

Possible Values: 
* `"Allow"` (default) - Action is Allowed
* `"Deny"` - Action is Denied

***

### .spec.scope.statements\[int\].resources

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.3/pkg/apis/permission/v1alpha1/policy/statement.go#L56)</sup>

This field is **required**

Resources defines the list of resources

***

### .spec.ttl

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.3/pkg/apis/permission/v1alpha1/token_spec.go#L54)</sup>

TTL Defines the TTL of the token.

Default Value: `1h`

