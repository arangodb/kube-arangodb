---
layout: page
parent: CRD reference
title: ArangoRoute V1Alpha1
---

# API Reference for ArangoRoute V1Alpha1

## Spec

### .spec.deployment

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/networking/v1alpha1/route_spec.go#L27)</sup>

DeploymentName specifies the ArangoDeployment object name

***

### .spec.destination.schema

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/networking/v1alpha1/route_spec_destination.go#L30)</sup>

Schema defines HTTP/S schema used for connection

***

### .spec.destination.service.checksum

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/shared/v1/object.go#L61)</sup>

UID keeps the information about object Checksum

***

### .spec.destination.service.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/shared/v1/object.go#L52)</sup>

Name of the object

***

### .spec.destination.service.namespace

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/shared/v1/object.go#L55)</sup>

Namespace of the object. Should default to the namespace of the parent object

***

### .spec.destination.service.port

Type: `intstr.IntOrString` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/networking/v1alpha1/route_spec_destination_service.go#L36)</sup>

Port defines Port or Port Name used as destination

Links:
* [Documentation](https://kubernetes.io/docs/tasks/administer-cluster/sysctl-cluster/)

***

### .spec.destination.service.uid

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/shared/v1/object.go#L58)</sup>

UID keeps the information about object UID

***

### .spec.destination.tls.insecure

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/networking/v1alpha1/route_spec_destination_tls.go#L25)</sup>

Insecure allows Insecure traffic

***

### .spec.route.path

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/networking/v1alpha1/route_spec_route.go#L29)</sup>

Path specifies the Path route

## Status

### .status.conditions

Type: `api.Conditions` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/networking/v1alpha1/route_status.go#L31)</sup>

Conditions specific to the entire extension

***

### .status.deployment.checksum

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/shared/v1/object.go#L61)</sup>

UID keeps the information about object Checksum

***

### .status.deployment.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/shared/v1/object.go#L52)</sup>

Name of the object

***

### .status.deployment.namespace

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/shared/v1/object.go#L55)</sup>

Namespace of the object. Should default to the namespace of the parent object

***

### .status.deployment.uid

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/shared/v1/object.go#L58)</sup>

UID keeps the information about object UID

***

### .status.targets\[int\].tls.insecure

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/networking/v1alpha1/route_status_target_tls.go#L27)</sup>

Insecure allows Insecure traffic

***

### .status.targets\[int\].url

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.42/pkg/apis/networking/v1alpha1/route_status_target.go#L34)</sup>

