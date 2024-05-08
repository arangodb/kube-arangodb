---
layout: page
parent: CRD reference
title: GraphAnalyticsEngine V1Alpha1
---

# API Reference for GraphAnalyticsEngine V1Alpha1

## Spec

### .spec.deploymentName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.40/pkg/apis/analytics/v1alpha1/gae_spec.go#L30)</sup>

DeploymentName define deployment name used in the object. Immutable

## Status

### .status.conditions

Type: `api.Conditions` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.40/pkg/apis/analytics/v1alpha1/gae_status.go#L31)</sup>

Conditions specific to the entire extension

***

### .status.deployment.checksum

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.40/pkg/apis/shared/v1/object.go#L61)</sup>

UID keeps the information about object Checksum

***

### .status.deployment.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.40/pkg/apis/shared/v1/object.go#L52)</sup>

Name of the object

***

### .status.deployment.namespace

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.40/pkg/apis/shared/v1/object.go#L55)</sup>

Namespace of the object. Should default to the namespace of the parent object

***

### .status.deployment.uid

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.40/pkg/apis/shared/v1/object.go#L58)</sup>

UID keeps the information about object UID

