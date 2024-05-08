---
layout: page
parent: CRD reference
title: GraphAnalyticsEngine V1Alpha1
---

# API Reference for GraphAnalyticsEngine V1Alpha1

## Spec

### .spec.deployment.port

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.40/pkg/apis/analytics/v1alpha1/gae_spec_deployment.go#L42)</sup>

Port defines on which port the container will be listening for connections

***

### .spec.deployment.service.type

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.40/pkg/apis/analytics/v1alpha1/gae_spec_deployment_service.go#L38)</sup>

Type determines how the Service is exposed

Links:
* [Kubernetes Documentation](https://kubernetes.io/docs/concepts/services-networking/service/#publishing-services-service-types)

Possible Values: 
* `"ClusterIP"` (default) - service will only be accessible inside the cluster, via the cluster IP
* `"NodePort"` - service will be exposed on one port of every node, in addition to 'ClusterIP' type
* `"LoadBalancer"` - service will be exposed via an external load balancer (if the cloud provider supports it), in addition to 'NodePort' type
* `"ExternalName"` - service consists of only a reference to an external name that kubedns or equivalent will return as a CNAME record, with no exposing or proxying of any pods involved
* `"None"` - service is not created

***

### .spec.deployment.tls.altNames

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.40/pkg/apis/shared/v1/tls.go#L28)</sup>

AltNames define TLS AltNames used when TLS on the ArangoDB is enabled

***

### .spec.deployment.tls.enabled

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.40/pkg/apis/shared/v1/tls.go#L25)</sup>

Enabled define if TLS Should be enabled. If is not set then default is taken from ArangoDeployment settings

***

### .spec.deploymentName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.40/pkg/apis/analytics/v1alpha1/gae_spec.go#L30)</sup>

DeploymentName define deployment name used in the object. Immutable

## Status

### .status.arangoDB.deployment.checksum

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.40/pkg/apis/shared/v1/object.go#L61)</sup>

UID keeps the information about object Checksum

***

### .status.arangoDB.deployment.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.40/pkg/apis/shared/v1/object.go#L52)</sup>

Name of the object

***

### .status.arangoDB.deployment.namespace

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.40/pkg/apis/shared/v1/object.go#L55)</sup>

Namespace of the object. Should default to the namespace of the parent object

***

### .status.arangoDB.deployment.uid

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.40/pkg/apis/shared/v1/object.go#L58)</sup>

UID keeps the information about object UID

***

### .status.arangoDB.secret.checksum

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.40/pkg/apis/shared/v1/object.go#L61)</sup>

UID keeps the information about object Checksum

***

### .status.arangoDB.secret.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.40/pkg/apis/shared/v1/object.go#L52)</sup>

Name of the object

***

### .status.arangoDB.secret.namespace

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.40/pkg/apis/shared/v1/object.go#L55)</sup>

Namespace of the object. Should default to the namespace of the parent object

***

### .status.arangoDB.secret.uid

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.40/pkg/apis/shared/v1/object.go#L58)</sup>

UID keeps the information about object UID

***

### .status.arangoDB.tls.checksum

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.40/pkg/apis/shared/v1/object.go#L61)</sup>

UID keeps the information about object Checksum

***

### .status.arangoDB.tls.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.40/pkg/apis/shared/v1/object.go#L52)</sup>

Name of the object

***

### .status.arangoDB.tls.namespace

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.40/pkg/apis/shared/v1/object.go#L55)</sup>

Namespace of the object. Should default to the namespace of the parent object

***

### .status.arangoDB.tls.uid

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.40/pkg/apis/shared/v1/object.go#L58)</sup>

UID keeps the information about object UID

***

### .status.conditions

Type: `api.Conditions` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.40/pkg/apis/analytics/v1alpha1/gae_status.go#L30)</sup>

Conditions specific to the entire extension

***

### .status.reconciliation.service.checksum

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.40/pkg/apis/shared/v1/object.go#L61)</sup>

UID keeps the information about object Checksum

***

### .status.reconciliation.service.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.40/pkg/apis/shared/v1/object.go#L52)</sup>

Name of the object

***

### .status.reconciliation.service.namespace

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.40/pkg/apis/shared/v1/object.go#L55)</sup>

Namespace of the object. Should default to the namespace of the parent object

***

### .status.reconciliation.service.uid

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.40/pkg/apis/shared/v1/object.go#L58)</sup>

UID keeps the information about object UID

***

### .status.reconciliation.statefulSet.checksum

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.40/pkg/apis/shared/v1/object.go#L61)</sup>

UID keeps the information about object Checksum

***

### .status.reconciliation.statefulSet.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.40/pkg/apis/shared/v1/object.go#L52)</sup>

Name of the object

***

### .status.reconciliation.statefulSet.namespace

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.40/pkg/apis/shared/v1/object.go#L55)</sup>

Namespace of the object. Should default to the namespace of the parent object

***

### .status.reconciliation.statefulSet.uid

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.40/pkg/apis/shared/v1/object.go#L58)</sup>

UID keeps the information about object UID

