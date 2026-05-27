---
layout: page
parent: CRD reference
title: ArangoPlatformConnector V1Beta1
---

# API Reference for ArangoPlatformConnector V1Beta1

## Spec

### .spec.deployment.checksum

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.3/pkg/apis/shared/v1/object.go#L62)</sup>

UID keeps the information about object Checksum

***

### .spec.deployment.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.3/pkg/apis/shared/v1/object.go#L53)</sup>

This field is **required**

Name of the object

***

### .spec.deployment.namespace

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.3/pkg/apis/shared/v1/object.go#L56)</sup>

Namespace of the object. Should default to the namespace of the parent object

***

### .spec.deployment.uid

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.3/pkg/apis/shared/v1/object.go#L59)</sup>

UID keeps the information about object UID

***

### .spec.description

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.3/pkg/apis/platform/v1beta1/connector_spec.go#L53)</sup>

Description of what this connector does

***

### .spec.route.checksum

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.3/pkg/apis/shared/v1/object.go#L62)</sup>

UID keeps the information about object Checksum

***

### .spec.route.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.3/pkg/apis/shared/v1/object.go#L53)</sup>

This field is **required**

Name of the object

***

### .spec.route.namespace

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.3/pkg/apis/shared/v1/object.go#L56)</sup>

Namespace of the object. Should default to the namespace of the parent object

***

### .spec.route.uid

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.3/pkg/apis/shared/v1/object.go#L59)</sup>

UID keeps the information about object UID

***

### .spec.schema

Type: `Object` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.3/pkg/apis/platform/v1beta1/connector_spec.go#L63)</sup>

Schema defines the JSON Schema for the connector's input query.
AI tools use this to validate parameters before submitting jobs.
Uses the same format as CRD validation schemas.

Links:
* [Kubernetes JSON Schema docs](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/#specifying-a-structural-schema)

***

### .spec.tags

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.3/pkg/apis/platform/v1beta1/connector_spec.go#L56)</sup>

Tags for discovery and filtering (e.g. "database", "aql", "vector-search")

***

### .spec.type

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.3/pkg/apis/platform/v1beta1/connector_spec.go#L44)</sup>

Type defines the connector pattern type

Possible Values: 
* `"Active"` (default) - Connector actively polls for and processes jobs

***

### .spec.version

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.3/pkg/apis/platform/v1beta1/connector_spec.go#L66)</sup>

Version of the connector

