---
layout: page
parent: CRD reference
title: ArangoPlatformConnector V1Beta1
---

# API Reference for ArangoPlatformConnector V1Beta1

## Spec

### .spec.deployment.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.3/pkg/apis/shared/v1/object.go#L53)</sup>

This field is **required**

Name of the object

***

### .spec.description

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.3/pkg/apis/platform/v1beta1/connector_spec.go#L69)</sup>

Description is a human-readable text explaining what this connector does.
Shown to AI tools via /_inventory for discovery. Set by the user.
Example: "Execute AQL queries on ArangoDB"

***

### .spec.route.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.3/pkg/apis/shared/v1/object.go#L53)</sup>

This field is **required**

Name of the object

***

### .spec.schema

Type: `Object` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.3/pkg/apis/platform/v1beta1/connector_spec.go#L83)</sup>

Schema defines the JSON Schema that describes the expected format of the query
field when submitting jobs to this connector. AI tools read this from /_inventory
to validate input before creating a job. Set by the user.
Uses the standard Kubernetes JSONSchemaProps format (same as CRD validation schemas).
The platform validates submitted job queries against this schema.

Links:
* [Kubernetes JSON Schema docs](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/#specifying-a-structural-schema)

***

### .spec.tags

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.3/pkg/apis/platform/v1beta1/connector_spec.go#L74)</sup>

Tags are labels used by AI tools to discover and filter connectors via /_inventory.
Set by the user. Use lowercase, descriptive terms.
Example: ["database", "aql", "query"]

***

### .spec.type

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.3/pkg/apis/platform/v1beta1/connector_spec.go#L47)</sup>

Type defines the connector execution pattern.
Currently only "Active" is supported — the connector runs as a long-lived process
that polls for pending jobs and processes them sequentially.
Set by the user when creating the connector. Defaults to "Active" if omitted.

Possible Values: 
* `"Active"` (default) - Connector actively polls for and processes jobs

***

### .spec.version

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.3/pkg/apis/platform/v1beta1/connector_spec.go#L88)</sup>

Version is the version string of the connector implementation.
Set by the user. Shown to AI tools via /_inventory. No format enforced,
but semantic versioning (e.g. "1.0.0") is recommended.

