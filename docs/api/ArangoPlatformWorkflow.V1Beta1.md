---
layout: page
parent: CRD reference
title: ArangoPlatformWorkflow V1Beta1
---

# API Reference for ArangoPlatformWorkflow V1Beta1

## Spec

### .spec.chart.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.5/pkg/apis/shared/v1/object.go#L53)</sup>

This field is **required**

Name of the object

***

### .spec.deployment.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.5/pkg/apis/shared/v1/object.go#L53)</sup>

This field is **required**

Name of the object

***

### .spec.hibernate

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.5/pkg/apis/platform/v1beta1/workflow_spec.go#L52)</sup>

Hibernate, when true, injects the hibernated flag into the chart values (arangodb_platform.hibernated)
so the chart can render its hibernated form.

Default Value: `false`

***

### .spec.install.timeout

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.5/pkg/apis/platform/v1beta1/workflow_spec_install.go#L35)</sup>

Timeout defines the install timeout

Default Value: `20m`

***

### .spec.upgrade.maxHistory

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.5/pkg/apis/platform/v1beta1/workflow_spec_upgrade.go#L39)</sup>

MaxHistory defines the max history

Default Value: `10`

***

### .spec.upgrade.timeout

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.5/pkg/apis/platform/v1beta1/workflow_spec_upgrade.go#L35)</sup>

Timeout defines the upgrade timeout

Default Value: `20m`

***

### .spec.values

Type: `Object` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.5/pkg/apis/platform/v1beta1/workflow_spec.go#L47)</sup>

Values keeps the values of the Workflow

