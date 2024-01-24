---
layout: page
parent: CRD reference
title: ArangoMLStorage V1Alpha1
---

# API Reference for ArangoMLStorage V1Alpha1

## Spec

### .spec.backend.s3.allowInsecure

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.36/pkg/apis/ml/v1alpha1/storage_spec_backend_s3.go#L40)</sup>

AllowInsecure if set to true, the Endpoint certificates won't be checked

Default Value: `false`

***

### .spec.backend.s3.caSecret.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.36/pkg/apis/shared/v1/object.go#L46)</sup>

Name of the object

***

### .spec.backend.s3.caSecret.namespace

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.36/pkg/apis/shared/v1/object.go#L49)</sup>

Namespace of the object. Should default to the namespace of the parent object

***

### .spec.backend.s3.caSecret.uid

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.36/pkg/apis/shared/v1/object.go#L52)</sup>

UID keeps the information about object UID

***

### .spec.backend.s3.credentialsSecret.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.36/pkg/apis/shared/v1/object.go#L46)</sup>

Name of the object

***

### .spec.backend.s3.credentialsSecret.namespace

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.36/pkg/apis/shared/v1/object.go#L49)</sup>

Namespace of the object. Should default to the namespace of the parent object

***

### .spec.backend.s3.credentialsSecret.uid

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.36/pkg/apis/shared/v1/object.go#L52)</sup>

UID keeps the information about object UID

***

### .spec.backend.s3.endpoint

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.36/pkg/apis/ml/v1alpha1/storage_spec_backend_s3.go#L34)</sup>

Endpoint specifies the S3 API-compatible endpoint which implements storage
Required

***

### .spec.backend.s3.region

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.36/pkg/apis/ml/v1alpha1/storage_spec_backend_s3.go#L49)</sup>

Region defines the availability zone name.

Default Value: `""`

***

### .spec.bucketName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.36/pkg/apis/ml/v1alpha1/storage_spec.go#L30)</sup>

BucketName specifies the name of the bucket
Required

***

### .spec.bucketPath

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.36/pkg/apis/ml/v1alpha1/storage_spec.go#L34)</sup>

BucketPath specifies the path within the bucket

Default Value: `/`

***

### .spec.mode.sidecar.controllerListenPort

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.36/pkg/apis/ml/v1alpha1/storage_spec_mode_sidecar.go#L36)</sup>

ControllerListenPort defines on which port the sidecar container will be listening for controller requests

Default Value: `9202`

***

### .spec.mode.sidecar.env

Type: `core.EnvVar` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.36/pkg/apis/shared/v1/envs.go#L33)</sup>

Env keeps the information about environment variables provided to the container

Links:
* [Kubernetes Docs](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#envvar-v1-core)

***

### .spec.mode.sidecar.envFrom

Type: `core.EnvFromSource` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.36/pkg/apis/shared/v1/envs.go#L38)</sup>

EnvFrom keeps the information about environment variable sources provided to the container

Links:
* [Kubernetes Docs](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#envfromsource-v1-core)

***

### .spec.mode.sidecar.image

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.36/pkg/apis/shared/v1/image.go#L31)</sup>

Image define image details

***

### .spec.mode.sidecar.listenPort

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.36/pkg/apis/ml/v1alpha1/storage_spec_mode_sidecar.go#L32)</sup>

ListenPort defines on which port the sidecar container will be listening for connections

Default Value: `9201`

***

### .spec.mode.sidecar.pullPolicy

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.36/pkg/apis/shared/v1/image.go#L35)</sup>

PullPolicy define Image pull policy

Default Value: `IfNotPresent`

***

### .spec.mode.sidecar.pullSecrets

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.36/pkg/apis/shared/v1/image.go#L38)</sup>

PullSecrets define Secrets used to pull Image from registry

***

### .spec.mode.sidecar.resources

Type: `core.ResourceRequirements` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.36/pkg/apis/shared/v1/resources.go#L34)</sup>

Resources holds resource requests & limits for container

Links:
* [Documentation of core.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#resourcerequirements-v1-core)

***

### .spec.mode.sidecar.securityContext

Type: `core.SecurityContext` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.36/pkg/apis/shared/v1/security_container.go#L29)</sup>

PodSecurityContext holds pod-level security attributes and common container settings.

Links:
* [Kubernetes docs](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/)

## Status

### .status.conditions

Type: `api.Conditions` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.36/pkg/apis/ml/v1alpha1/storage_status.go#L28)</sup>

Conditions specific to the entire storage

