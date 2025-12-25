---
layout: page
parent: CRD reference
title: ArangoMLStorage V1Beta1
---

# API Reference for ArangoMLStorage V1Beta1

## Spec

### .spec.backend.s3.allowInsecure

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.3.4/pkg/apis/ml/v1beta1/storage_spec_backend_s3.go#L40)</sup>

AllowInsecure if set to true, the Endpoint certificates won't be checked

Default Value: `false`

***

### .spec.backend.s3.caSecret.checksum

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.3.4/pkg/apis/shared/v1/object.go#L62)</sup>

UID keeps the information about object Checksum

***

### .spec.backend.s3.caSecret.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.3.4/pkg/apis/shared/v1/object.go#L53)</sup>

This field is **required**

Name of the object

***

### .spec.backend.s3.caSecret.namespace

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.3.4/pkg/apis/shared/v1/object.go#L56)</sup>

Namespace of the object. Should default to the namespace of the parent object

***

### .spec.backend.s3.caSecret.uid

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.3.4/pkg/apis/shared/v1/object.go#L59)</sup>

UID keeps the information about object UID

***

### .spec.backend.s3.credentialsSecret.checksum

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.3.4/pkg/apis/shared/v1/object.go#L62)</sup>

UID keeps the information about object Checksum

***

### .spec.backend.s3.credentialsSecret.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.3.4/pkg/apis/shared/v1/object.go#L53)</sup>

This field is **required**

Name of the object

***

### .spec.backend.s3.credentialsSecret.namespace

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.3.4/pkg/apis/shared/v1/object.go#L56)</sup>

Namespace of the object. Should default to the namespace of the parent object

***

### .spec.backend.s3.credentialsSecret.uid

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.3.4/pkg/apis/shared/v1/object.go#L59)</sup>

UID keeps the information about object UID

***

### .spec.backend.s3.endpoint

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.3.4/pkg/apis/ml/v1beta1/storage_spec_backend_s3.go#L34)</sup>

This field is **required**

Endpoint specifies the S3 API-compatible endpoint which implements storage

***

### .spec.backend.s3.region

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.3.4/pkg/apis/ml/v1beta1/storage_spec_backend_s3.go#L49)</sup>

Region defines the availability zone name.

Default Value: `""`

***

### .spec.bucketName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.3.4/pkg/apis/ml/v1beta1/storage_spec.go#L30)</sup>

This field is **required**

BucketName specifies the name of the bucket

***

### .spec.bucketPath

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.3.4/pkg/apis/ml/v1beta1/storage_spec.go#L34)</sup>

BucketPath specifies the path within the bucket

Default Value: `/`

