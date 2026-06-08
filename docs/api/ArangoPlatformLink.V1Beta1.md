---
layout: page
parent: CRD reference
title: ArangoPlatformLink V1Beta1
---

# API Reference for ArangoPlatformLink V1Beta1

## Spec

### .spec.deployment.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.3/pkg/apis/shared/v1/object.go#L53)</sup>

This field is **required**

Name of the object

***

### .spec.route.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.3/pkg/apis/shared/v1/object.go#L53)</sup>

This field is **required**

Name of the object

***

### .spec.type

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.3/pkg/apis/platform/v1beta1/link_spec.go#L45)</sup>

Type defines the link execution pattern.
Currently only "Active" is supported — the link runs as a long-lived process
that polls for pending jobs and processes them sequentially.
Set by the user when creating the link. Defaults to "Active" if omitted.

Possible Values: 
* `"Active"` (default) - Link actively polls for and processes jobs

