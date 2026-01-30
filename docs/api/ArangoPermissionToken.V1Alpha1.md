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

### .spec.roles

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.1/pkg/apis/permission/v1alpha1/token_spec.go#L46)</sup>

Roles keeps the roles assigned to the token

***

### .spec.ttl

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.1/pkg/apis/permission/v1alpha1/token_spec.go#L51)</sup>

TTL Defines the TTL of the token.

Default Value: `1h`

