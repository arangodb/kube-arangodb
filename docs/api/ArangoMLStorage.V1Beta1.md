---
layout: page
parent: CRD reference
title: ArangoMLStorage V1Beta1
---

# API Reference for ArangoMLStorage V1Beta1

## Spec

### .spec.backend.s3.allowInsecure

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.47/pkg/apis/ml/v1beta1/storage_spec_backend_s3.go#L40)</sup>

AllowInsecure if set to true, the Endpoint certificates won't be checked

Default Value: `false`

***

### .spec.backend.s3.caSecret.checksum

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.47/pkg/apis/shared/v1/object.go#L61)</sup>

UID keeps the information about object Checksum

***

### .spec.backend.s3.caSecret.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.47/pkg/apis/shared/v1/object.go#L52)</sup>

Name of the object

***

### .spec.backend.s3.caSecret.namespace

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.47/pkg/apis/shared/v1/object.go#L55)</sup>

Namespace of the object. Should default to the namespace of the parent object

***

### .spec.backend.s3.caSecret.uid

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.47/pkg/apis/shared/v1/object.go#L58)</sup>

UID keeps the information about object UID

***

### .spec.backend.s3.credentialsSecret.checksum

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.47/pkg/apis/shared/v1/object.go#L61)</sup>

UID keeps the information about object Checksum

***

### .spec.backend.s3.credentialsSecret.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.47/pkg/apis/shared/v1/object.go#L52)</sup>

Name of the object

***

### .spec.backend.s3.credentialsSecret.namespace

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.47/pkg/apis/shared/v1/object.go#L55)</sup>

Namespace of the object. Should default to the namespace of the parent object

***

### .spec.backend.s3.credentialsSecret.uid

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.47/pkg/apis/shared/v1/object.go#L58)</sup>

UID keeps the information about object UID

***

### .spec.backend.s3.endpoint

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.47/pkg/apis/ml/v1beta1/storage_spec_backend_s3.go#L34)</sup>

Endpoint specifies the S3 API-compatible endpoint which implements storage
Required

***

### .spec.backend.s3.region

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.47/pkg/apis/ml/v1beta1/storage_spec_backend_s3.go#L49)</sup>

Region defines the availability zone name.

Default Value: `""`

***

### .spec.bucketName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.47/pkg/apis/ml/v1beta1/storage_spec.go#L30)</sup>

BucketName specifies the name of the bucket
Required

***

### .spec.bucketPath

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.47/pkg/apis/ml/v1beta1/storage_spec.go#L34)</sup>

BucketPath specifies the path within the bucket

Default Value: `/`

## Status

### .status.conditions

Type: `api.Conditions` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.47/pkg/apis/ml/v1beta1/storage_status.go#L28)</sup>

Conditions specific to the entire storage

