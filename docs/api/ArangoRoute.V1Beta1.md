---
layout: page
parent: CRD reference
title: ArangoRoute V1Beta1
---

# API Reference for ArangoRoute V1Beta1

## Spec

### .spec.deployment

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.3.2/pkg/apis/networking/v1beta1/route_spec.go#L28)</sup>

This field is **required**

Deployment specifies the ArangoDeployment object name

***

### .spec.destination.authentication.passMode

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.3.2/pkg/apis/networking/v1beta1/route_spec_destination_authentication.go#L32)</sup>

PassMode define authorization details pass mode when authorization was successful

Possible Values: 
* `"override"` (default) - Generates new token for the user
* `"pass"` - Pass token provided by the user
* `"remove"` - Removes authorization details from the request

***

### .spec.destination.authentication.type

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.3.2/pkg/apis/networking/v1beta1/route_spec_destination_authentication.go#L37)</sup>

Type of the authentication

Possible Values: 
* `"optional"` (default) - Authentication is header is validated and passed to the service. In case if is unauthorized, requests is still passed
* `"required"` - Authentication is header is validated and passed to the service. In case if is unauthorized, returns 403

***

### .spec.destination.endpoints.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.3.2/pkg/apis/shared/v1/object.go#L53)</sup>

This field is **required**

Name of the object

***

### .spec.destination.endpoints.namespace

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.3.2/pkg/apis/shared/v1/object.go#L56)</sup>

Namespace of the object. Should default to the namespace of the parent object

***

### .spec.destination.endpoints.port

Type: `intstr.IntOrString` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.3.2/pkg/apis/networking/v1beta1/route_spec_destination_endpoint.go#L39)</sup>

This field is **required**

Port defines Port or Port Name used as destination

***

### .spec.destination.path

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.3.2/pkg/apis/networking/v1beta1/route_spec_destination.go#L57)</sup>

Path defines service path used for overrides

***

### .spec.destination.protocol

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.3.2/pkg/apis/networking/v1beta1/route_spec_destination.go#L51)</sup>

Protocol defines http protocol used for the route

Possible Values: 
* `"http1"` (default) - HTTP 1.1 Protocol
* `"http2"` - HTTP 2 Protocol

***

### .spec.destination.redirect.code

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.3.2/pkg/apis/networking/v1beta1/route_spec_destination_redirect.go#L33)</sup>

Code the redirection response status code

Default Value: `307`

***

### .spec.destination.schema

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.3.2/pkg/apis/networking/v1beta1/route_spec_destination.go#L45)</sup>

Schema defines HTTP/S schema used for connection

Possible Values: 
* `"http"` (default) - HTTP Connection
* `"https"` - HTTPS Connection (HTTP with TLS)

***

### .spec.destination.service.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.3.2/pkg/apis/shared/v1/object.go#L53)</sup>

This field is **required**

Name of the object

***

### .spec.destination.service.namespace

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.3.2/pkg/apis/shared/v1/object.go#L56)</sup>

Namespace of the object. Should default to the namespace of the parent object

***

### .spec.destination.service.port

Type: `intstr.IntOrString` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.3.2/pkg/apis/networking/v1beta1/route_spec_destination_service.go#L38)</sup>

This field is **required**

Port defines Port or Port Name used as destination

***

### .spec.destination.timeout

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.3.2/pkg/apis/networking/v1beta1/route_spec_destination.go#L65)</sup>

Timeout specify the upstream request timeout

Default Value: `1m0s`

***

### .spec.destination.tls.insecure

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.3.2/pkg/apis/networking/v1beta1/route_spec_destination_tls.go#L25)</sup>

Insecure allows Insecure traffic

***

### .spec.options.upgrade\[int\].enabled

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.3.2/pkg/apis/networking/v1beta1/route_spec_options_upgrade.go#L50)</sup>

Enabled defines if upgrade option is enabled

***

### .spec.options.upgrade\[int\].type

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.3.2/pkg/apis/networking/v1beta1/route_spec_options_upgrade.go#L47)</sup>

Type defines type of the Upgrade

Possible Values: 
* `"websocket"` (default) - HTTP WebSocket Upgrade type

***

### .spec.route.path

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.3.2/pkg/apis/networking/v1beta1/route_spec_route.go#L29)</sup>

Path specifies the Path route

