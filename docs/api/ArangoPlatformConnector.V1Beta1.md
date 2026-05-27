---
layout: page
parent: CRD reference
title: ArangoPlatformConnector V1Beta1
---

# API Reference for ArangoPlatformConnector V1Beta1

## Spec

### .spec.description

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.3/pkg/apis/platform/v1beta1/connector_spec.go#L29)</sup>

Description of what this connector does

***

### .spec.schema

Type: `Object` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.3/pkg/apis/platform/v1beta1/connector_spec.go#L39)</sup>

Schema defines the JSON Schema for the connector's input query.
AI tools use this to validate parameters before submitting jobs.
Uses the same format as CRD validation schemas.

Links:
* [Kubernetes JSON Schema docs](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/#specifying-a-structural-schema)

***

### .spec.tags

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.3/pkg/apis/platform/v1beta1/connector_spec.go#L32)</sup>

Tags for discovery and filtering (e.g. "database", "aql", "vector-search")

***

### .spec.version

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.3/pkg/apis/platform/v1beta1/connector_spec.go#L42)</sup>

Version of the connector

