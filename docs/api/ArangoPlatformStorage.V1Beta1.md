---
layout: page
parent: CRD reference
title: ArangoPlatformStorage V1Beta1
---

# API Reference for ArangoPlatformStorage V1Beta1

## Spec

### .spec.backend.azureBlobStorage.accountName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.3.4/pkg/apis/platform/v1beta1/storage_spec_backend_abs.go#L39)</sup>

This field is **required**

AccountName specifies the Azure Storage AccountName
used in format https://<account>.blob.core.windows.net/

***

### .spec.backend.azureBlobStorage.bucketName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.3.4/pkg/apis/platform/v1beta1/storage_spec_backend_abs.go#L46)</sup>

This field is **required**

BucketName specifies the name of the bucket

***

### .spec.backend.azureBlobStorage.bucketPath

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.3.4/pkg/apis/platform/v1beta1/storage_spec_backend_abs.go#L50)</sup>

BucketPath specifies the Prefix within the bucket

***

### .spec.backend.azureBlobStorage.credentialsSecret.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.3.4/pkg/apis/shared/v1/object.go#L53)</sup>

This field is **required**

Name of the object

***

### .spec.backend.azureBlobStorage.endpoint

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.3.4/pkg/apis/platform/v1beta1/storage_spec_backend_abs.go#L42)</sup>

Endpoint specifies the Azure Storage custom endpoint

***

### .spec.backend.azureBlobStorage.tenantID

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.3.4/pkg/apis/platform/v1beta1/storage_spec_backend_abs.go#L34)</sup>

This field is **required**

TenantID specifies the Azure TenantID

***

### .spec.backend.gcs.bucketName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.3.4/pkg/apis/platform/v1beta1/storage_spec_backend_gcs.go#L35)</sup>

This field is **required**

BucketName specifies the name of the bucket

***

### .spec.backend.gcs.bucketPath

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.3.4/pkg/apis/platform/v1beta1/storage_spec_backend_gcs.go#L38)</sup>

BucketPath specifies the Prefix within the bucket

***

### .spec.backend.gcs.credentialsSecret.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.3.4/pkg/apis/shared/v1/object.go#L53)</sup>

This field is **required**

Name of the object

***

### .spec.backend.gcs.projectID

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.3.4/pkg/apis/platform/v1beta1/storage_spec_backend_gcs.go#L32)</sup>

This field is **required**

ProjectID specifies the GCP ProjectID

***

### .spec.backend.s3.allowInsecure

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.3.4/pkg/apis/platform/v1beta1/storage_spec_backend_s3.go#L49)</sup>

AllowInsecure if set to true, the Endpoint certificates won't be checked

Default Value: `false`

***

### .spec.backend.s3.bucketName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.3.4/pkg/apis/platform/v1beta1/storage_spec_backend_s3.go#L34)</sup>

This field is **required**

BucketName specifies the name of the bucket

***

### .spec.backend.s3.bucketPath

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.3.4/pkg/apis/platform/v1beta1/storage_spec_backend_s3.go#L37)</sup>

BucketPath specifies the Prefix within the bucket

***

### .spec.backend.s3.caSecret.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.3.4/pkg/apis/shared/v1/object.go#L53)</sup>

This field is **required**

Name of the object

***

### .spec.backend.s3.credentialsSecret.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.3.4/pkg/apis/shared/v1/object.go#L53)</sup>

This field is **required**

Name of the object

***

### .spec.backend.s3.endpoint

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.3.4/pkg/apis/platform/v1beta1/storage_spec_backend_s3.go#L40)</sup>

This field is **required**

Endpoint specifies the S3 API-compatible endpoint which implements storage

***

### .spec.backend.s3.region

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.3.4/pkg/apis/platform/v1beta1/storage_spec_backend_s3.go#L61)</sup>

Region defines the availability zone name.

Default Value: `""`

