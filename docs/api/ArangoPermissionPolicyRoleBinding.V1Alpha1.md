---
layout: page
parent: CRD reference
title: ArangoPermissionPolicyRoleBinding V1Alpha1
---

# API Reference for ArangoPermissionPolicyRoleBinding V1Alpha1

## Spec

### .spec.deployment.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.3/pkg/apis/shared/v1/object.go#L53)</sup>

This field is **required**

Name of the object

***

### .spec.policy.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.3/pkg/apis/permission/v1alpha1/binding_ref.go#L33)</sup>

Name references an ArangoPermission CRD by name. The operator resolves it to the sidecar name.

***

### .spec.role.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.3/pkg/apis/permission/v1alpha1/binding_ref.go#L33)</sup>

Name references an ArangoPermission CRD by name. The operator resolves it to the sidecar name.

