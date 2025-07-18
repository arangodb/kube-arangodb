---
layout: page
parent: CRD reference
title: ArangoRoute V1Alpha1
---

# API Reference for ArangoRoute V1Alpha1

## Spec

### .spec.deployment

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.50/pkg/apis/networking/v1alpha1/route_spec.go#L28)</sup>

This field is **required**

Deployment specifies the ArangoDeployment object name

***

### .spec.destination.authentication.passMode

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.50/pkg/apis/networking/v1alpha1/route_spec_destination_authentication.go#L32)</sup>

PassMode define authorization details pass mode when authorization was successful

Possible Values: 
* `"override"` (default) - Generates new token for the user
* `"pass"` - Pass token provided by the user
* `"remove"` - Removes authorization details from the request

***

### .spec.destination.authentication.type

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.50/pkg/apis/networking/v1alpha1/route_spec_destination_authentication.go#L37)</sup>

Type of the authentication

Possible Values: 
* `"optional"` (default) - Authentication is header is validated and passed to the service. In case if is unauthorized, requests is still passed
* `"required"` - Authentication is header is validated and passed to the service. In case if is unauthorized, returns 403

***

### .spec.destination.endpoints.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.50/pkg/apis/shared/v1/object.go#L53)</sup>

This field is **required**

Name of the object

***

### .spec.destination.endpoints.namespace

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.50/pkg/apis/shared/v1/object.go#L56)</sup>

Namespace of the object. Should default to the namespace of the parent object

***

### .spec.destination.endpoints.port

Type: `intstr.IntOrString` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.50/pkg/apis/networking/v1alpha1/route_spec_destination_endpoint.go#L39)</sup>

This field is **required**

Port defines Port or Port Name used as destination

***

### .spec.destination.path

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.50/pkg/apis/networking/v1alpha1/route_spec_destination.go#L52)</sup>

Path defines service path used for overrides

***

### .spec.destination.protocol

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.50/pkg/apis/networking/v1alpha1/route_spec_destination.go#L46)</sup>

Protocol defines http protocol used for the route

Possible Values: 
* `"http1"` (default) - HTTP 1.1 Protocol
* `"http2"` - HTTP 2 Protocol

***

### .spec.destination.schema

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.50/pkg/apis/networking/v1alpha1/route_spec_destination.go#L41)</sup>

Schema defines HTTP/S schema used for connection

Possible Values: 
* `"http"` (default) - HTTP Connection
* `"https"` - HTTPS Connection (HTTP with TLS)

***

### .spec.destination.service.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.50/pkg/apis/shared/v1/object.go#L53)</sup>

This field is **required**

Name of the object

***

### .spec.destination.service.namespace

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.50/pkg/apis/shared/v1/object.go#L56)</sup>

Namespace of the object. Should default to the namespace of the parent object

***

### .spec.destination.service.port

Type: `intstr.IntOrString` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.50/pkg/apis/networking/v1alpha1/route_spec_destination_service.go#L38)</sup>

This field is **required**

Port defines Port or Port Name used as destination

***

### .spec.destination.timeout

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.50/pkg/apis/networking/v1alpha1/route_spec_destination.go#L60)</sup>

Timeout specify the upstream request timeout

Default Value: `1m0s`

***

### .spec.destination.tls.insecure

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.50/pkg/apis/networking/v1alpha1/route_spec_destination_tls.go#L25)</sup>

Insecure allows Insecure traffic

***

### .spec.options.upgrade\[int\].enabled

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.50/pkg/apis/networking/v1alpha1/route_spec_options_upgrade.go#L50)</sup>

Enabled defines if upgrade option is enabled

***

### .spec.options.upgrade\[int\].type

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.50/pkg/apis/networking/v1alpha1/route_spec_options_upgrade.go#L47)</sup>

Type defines type of the Upgrade

Possible Values: 
* `"websocket"` (default) - HTTP WebSocket Upgrade type

***

### .spec.route.path

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.50/pkg/apis/networking/v1alpha1/route_spec_route.go#L29)</sup>

Path specifies the Path route

## Status

### .status.conditions

Type: `api.Conditions` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.50/pkg/apis/networking/v1alpha1/route_status.go#L31)</sup>

Conditions specific to the entire extension

***

### .status.deployment.checksum

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.50/pkg/apis/shared/v1/object.go#L62)</sup>

UID keeps the information about object Checksum

***

### .status.deployment.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.50/pkg/apis/shared/v1/object.go#L53)</sup>

This field is **required**

Name of the object

***

### .status.deployment.namespace

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.50/pkg/apis/shared/v1/object.go#L56)</sup>

Namespace of the object. Should default to the namespace of the parent object

***

### .status.deployment.uid

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.50/pkg/apis/shared/v1/object.go#L59)</sup>

UID keeps the information about object UID

***

### .status.target.authentication.passMode

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.50/pkg/apis/networking/v1alpha1/route_status_target_authentication.go#L27)</sup>

***

### .status.target.authentication.type

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.50/pkg/apis/networking/v1alpha1/route_status_target_authentication.go#L26)</sup>

***

### .status.target.destinations\[int\].host

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.50/pkg/apis/networking/v1alpha1/route_status_target_destination.go#L38)</sup>

***

### .status.target.destinations\[int\].port

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.50/pkg/apis/networking/v1alpha1/route_status_target_destination.go#L39)</sup>

***

### .status.target.options.upgrade\[int\].enabled

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.50/pkg/apis/networking/v1alpha1/route_status_target_options_upgrade.go#L43)</sup>

Enabled defines if upgrade option is enabled

***

### .status.target.options.upgrade\[int\].type

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.50/pkg/apis/networking/v1alpha1/route_status_target_options_upgrade.go#L40)</sup>

Type defines type of the Upgrade

Possible Values: 
* `"websocket"` (default) - HTTP WebSocket Upgrade type

***

### .status.target.path

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.50/pkg/apis/networking/v1alpha1/route_status_target.go#L51)</sup>

Path specifies request path override

***

### .status.target.protocol

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.50/pkg/apis/networking/v1alpha1/route_status_target.go#L42)</sup>

Protocol defines http protocol used for the route

***

### .status.target.route.path

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.50/pkg/apis/networking/v1alpha1/route_status_target_route.go#L29)</sup>

Path specifies the Path route

***

### .status.target.timeout

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.50/pkg/apis/networking/v1alpha1/route_status_target.go#L57)</sup>

Timeout specify the upstream request timeout

***

### .status.target.tls.insecure

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.50/pkg/apis/networking/v1alpha1/route_status_target_tls.go#L27)</sup>

Insecure allows Insecure traffic

***

### .status.target.type

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.50/pkg/apis/networking/v1alpha1/route_status_target.go#L36)</sup>

Type define destination type

