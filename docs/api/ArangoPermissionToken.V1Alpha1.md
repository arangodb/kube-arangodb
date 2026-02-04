---
layout: page
parent: CRD reference
title: ArangoPermissionToken V1Alpha1
---

# API Reference for ArangoPermissionToken V1Alpha1

## Spec

### .spec.deployment.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.1/pkg/apis/shared/v1/object.go#L53)</sup>

This field is **required**

Name of the object

***

### .spec.policy.statements\[int\].actions

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.1/pkg/apis/permission/v1alpha1/policy/statement.go#L44)</sup>

This field is **required**

Actions defines the list of actions.
Action needs to be defined in format `<namespace>:<name>`

***

### .spec.policy.statements\[int\].effect

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.1/pkg/apis/permission/v1alpha1/policy/statement.go#L39)</sup>

This field is **required**

Effect defines the statement effect.

Possible Values: 
* `"Allow"` (default) - Action is Allowed
* `"Deny"` - Action is Denied

***

### .spec.policy.statements\[int\].resources

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.1/pkg/apis/permission/v1alpha1/policy/statement.go#L48)</sup>

This field is **required**

Resources defines the list of resources

***

### .spec.roles

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.1/pkg/apis/permission/v1alpha1/token_spec.go#L47)</sup>

Roles keeps the roles assigned to the token

***

### .spec.ttl

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.1/pkg/apis/permission/v1alpha1/token_spec.go#L52)</sup>

TTL Defines the TTL of the token.

Default Value: `1h`

