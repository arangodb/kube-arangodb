---
layout: page
parent: CRD reference
title: ArangoPlatformStorage V1Alpha1
---

# API Reference for ArangoPlatformStorage V1Alpha1

## Spec

### .spec.backend.s3.allowInsecure

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.46/pkg/apis/platform/v1alpha1/storage_spec_backend_s3.go#L46)</sup>

AllowInsecure if set to true, the Endpoint certificates won't be checked

Default Value: `false`

***

### .spec.backend.s3.bucketName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.46/pkg/apis/platform/v1alpha1/storage_spec_backend_s3.go#L34)</sup>

BucketName specifies the name of the bucket
Required

***

### .spec.backend.s3.bucketPath

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.46/pkg/apis/platform/v1alpha1/storage_spec_backend_s3.go#L37)</sup>

BucketPath specifies the Prefix within the bucket

***

### .spec.backend.s3.caSecret.checksum

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.46/pkg/apis/shared/v1/object.go#L61)</sup>

UID keeps the information about object Checksum

***

### .spec.backend.s3.caSecret.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.46/pkg/apis/shared/v1/object.go#L52)</sup>

Name of the object

***

### .spec.backend.s3.caSecret.namespace

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.46/pkg/apis/shared/v1/object.go#L55)</sup>

Namespace of the object. Should default to the namespace of the parent object

***

### .spec.backend.s3.caSecret.uid

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.46/pkg/apis/shared/v1/object.go#L58)</sup>

UID keeps the information about object UID

***

### .spec.backend.s3.credentialsSecret.checksum

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.46/pkg/apis/shared/v1/object.go#L61)</sup>

UID keeps the information about object Checksum

***

### .spec.backend.s3.credentialsSecret.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.46/pkg/apis/shared/v1/object.go#L52)</sup>

Name of the object

***

### .spec.backend.s3.credentialsSecret.namespace

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.46/pkg/apis/shared/v1/object.go#L55)</sup>

Namespace of the object. Should default to the namespace of the parent object

***

### .spec.backend.s3.credentialsSecret.uid

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.46/pkg/apis/shared/v1/object.go#L58)</sup>

UID keeps the information about object UID

***

### .spec.backend.s3.endpoint

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.46/pkg/apis/platform/v1alpha1/storage_spec_backend_s3.go#L40)</sup>

Endpoint specifies the S3 API-compatible endpoint which implements storage
Required

***

### .spec.backend.s3.region

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.46/pkg/apis/platform/v1alpha1/storage_spec_backend_s3.go#L55)</sup>

Region defines the availability zone name.

Default Value: `""`

