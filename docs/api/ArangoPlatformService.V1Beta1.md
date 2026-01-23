---
layout: page
parent: CRD reference
title: ArangoPlatformService V1Beta1
---

# API Reference for ArangoPlatformService V1Beta1

## Spec

### .spec.chart.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.0/pkg/apis/shared/v1/object.go#L53)</sup>

This field is **required**

Name of the object

***

### .spec.deployment.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.0/pkg/apis/shared/v1/object.go#L53)</sup>

This field is **required**

Name of the object

***

### .spec.install.timeout

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.0/pkg/apis/platform/v1beta1/service_spec_install.go#L33)</sup>

Timeout defines the upgrade timeout

Default Value: `20m`

***

### .spec.upgrade.maxHistory

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.0/pkg/apis/platform/v1beta1/service_spec_upgrade.go#L37)</sup>

MaxHistory defines the max history

Default Value: `10`

***

### .spec.upgrade.timeout

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.0/pkg/apis/platform/v1beta1/service_spec_upgrade.go#L33)</sup>

Timeout defines the upgrade timeout

Default Value: `20m`

***

### .spec.values

Type: `Object` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.4.0/pkg/apis/platform/v1beta1/service_spec.go#L46)</sup>

Values keeps the values of the Service

