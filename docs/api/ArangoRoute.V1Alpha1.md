---
layout: page
parent: CRD reference
title: ArangoRoute V1Alpha1
---

# API Reference for ArangoRoute V1Alpha1

## Spec

### .spec.deployment

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.43/pkg/apis/networking/v1alpha1/route_spec.go#L27)</sup>

Deployment specifies the ArangoDeployment object name

***

### .spec.destination.authentication.passMode

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.43/pkg/apis/networking/v1alpha1/route_spec_destination_authentication.go#L32)</sup>

PassMode define authorization details pass mode when authorization was successful

Possible Values: 
* `"override"` (default) - Generates new token for the user
* `"pass"` - Pass token provided by the user
* `"remove"` - Removes authorization details from the request

***

### .spec.destination.authentication.type

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.43/pkg/apis/networking/v1alpha1/route_spec_destination_authentication.go#L37)</sup>

Type of the authentication

Possible Values: 
* `"optional"` (default) - Authentication is header is validated and passed to the service. In case if is unauthorized, requests is still passed
* `"required"` - Authentication is header is validated and passed to the service. In case if is unauthorized, returns 403

***

### .spec.destination.endpoints.checksum

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.43/pkg/apis/shared/v1/object.go#L61)</sup>

UID keeps the information about object Checksum

***

### .spec.destination.endpoints.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.43/pkg/apis/shared/v1/object.go#L52)</sup>

Name of the object

***

### .spec.destination.endpoints.namespace

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.43/pkg/apis/shared/v1/object.go#L55)</sup>

Namespace of the object. Should default to the namespace of the parent object

***

### .spec.destination.endpoints.port

Type: `intstr.IntOrString` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.43/pkg/apis/networking/v1alpha1/route_spec_destination_endpoint.go#L36)</sup>

Port defines Port or Port Name used as destination

***

### .spec.destination.endpoints.uid

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.43/pkg/apis/shared/v1/object.go#L58)</sup>

UID keeps the information about object UID

***

### .spec.destination.path

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.43/pkg/apis/networking/v1alpha1/route_spec_destination.go#L39)</sup>

Path defines service path used for overrides

***

### .spec.destination.schema

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.43/pkg/apis/networking/v1alpha1/route_spec_destination.go#L33)</sup>

Schema defines HTTP/S schema used for connection

***

### .spec.destination.service.checksum

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.43/pkg/apis/shared/v1/object.go#L61)</sup>

UID keeps the information about object Checksum

***

### .spec.destination.service.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.43/pkg/apis/shared/v1/object.go#L52)</sup>

Name of the object

***

### .spec.destination.service.namespace

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.43/pkg/apis/shared/v1/object.go#L55)</sup>

Namespace of the object. Should default to the namespace of the parent object

***

### .spec.destination.service.port

Type: `intstr.IntOrString` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.43/pkg/apis/networking/v1alpha1/route_spec_destination_service.go#L35)</sup>

Port defines Port or Port Name used as destination

***

### .spec.destination.service.uid

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.43/pkg/apis/shared/v1/object.go#L58)</sup>

UID keeps the information about object UID

***

### .spec.destination.tls.insecure

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.43/pkg/apis/networking/v1alpha1/route_spec_destination_tls.go#L25)</sup>

Insecure allows Insecure traffic

***

### .spec.route.path

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.43/pkg/apis/networking/v1alpha1/route_spec_route.go#L29)</sup>

Path specifies the Path route

## Status

### .status.conditions

Type: `api.Conditions` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.43/pkg/apis/networking/v1alpha1/route_status.go#L31)</sup>

Conditions specific to the entire extension

***

### .status.deployment.checksum

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.43/pkg/apis/shared/v1/object.go#L61)</sup>

UID keeps the information about object Checksum

***

### .status.deployment.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.43/pkg/apis/shared/v1/object.go#L52)</sup>

Name of the object

***

### .status.deployment.namespace

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.43/pkg/apis/shared/v1/object.go#L55)</sup>

Namespace of the object. Should default to the namespace of the parent object

***

### .status.deployment.uid

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.43/pkg/apis/shared/v1/object.go#L58)</sup>

UID keeps the information about object UID

***

### .status.target.authentication.passMode

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.43/pkg/apis/networking/v1alpha1/route_status_target_authentication.go#L27)</sup>

***

### .status.target.authentication.type

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.43/pkg/apis/networking/v1alpha1/route_status_target_authentication.go#L26)</sup>

***

### .status.target.destinations\[int\].host

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.43/pkg/apis/networking/v1alpha1/route_status_target_destination.go#L38)</sup>

***

### .status.target.destinations\[int\].port

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.43/pkg/apis/networking/v1alpha1/route_status_target_destination.go#L39)</sup>

***

### .status.target.path

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.43/pkg/apis/networking/v1alpha1/route_status_target.go#L43)</sup>

Path specifies request path override

***

### .status.target.TLS.insecure

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.43/pkg/apis/networking/v1alpha1/route_status_target_tls.go#L27)</sup>

Insecure allows Insecure traffic

***

### .status.target.type

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.43/pkg/apis/networking/v1alpha1/route_status_target.go#L34)</sup>

Type define destination type

