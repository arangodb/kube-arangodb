---
layout: page
parent: CRD reference
title: ArangoPermissionPolicyRoleBinding V1Alpha1
---

# API Reference for ArangoPermissionPolicyRoleBinding V1Alpha1

## Spec

### .spec.deployment.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.4/pkg/apis/shared/v1/object.go#L53)</sup>

This field is **required**

Name of the object

***

### .spec.policy.direct

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.4/pkg/apis/permission/v1alpha1/binding_ref.go#L40)</sup>

Direct references an existing authorization object (role or policy) by its exact name, without
a backing ArangoPermission CRD - e.g. an operator-managed predefined role
"managed:predefined:coredb-reader". The value is used as-is. Exactly one of Name or Direct
must be set.

***

### .spec.policy.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.4/pkg/apis/permission/v1alpha1/binding_ref.go#L34)</sup>

Name references an ArangoPermission CRD by name. The operator resolves it to the sidecar name.

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

