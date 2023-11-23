# API Reference for ArangoMLStorage V1Alpha1

## Spec

### .spec.listenPort

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/ml/v1alpha1/storage_spec.go#L32)</sup>

ListenPort defines on which port the sidecar container will be listening for connections

Default Value: `9201`

***

### .spec.resources

Type: `core.ResourceRequirements` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/ml/v1alpha1/storage_spec.go#L37)</sup>

Resources holds resource requests & limits for container running the S3 proxy

Links:
* [Documentation of core.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#resourcerequirements-v1-core)

***

### .spec.s3.bucketName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/ml/v1alpha1/storage_s3_spec.go#L39)</sup>

BucketName specifies the name of the bucket
Required

***

### .spec.s3.credentialsSecret

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/ml/v1alpha1/storage_s3_spec.go#L42)</sup>

CredentialsSecretName specifies the name of the secret containing AccessKey and SecretKey for S3 API authorization
Required

***

### .spec.s3.disableSSL

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/ml/v1alpha1/storage_s3_spec.go#L33)</sup>

DisableSSL if set to true, no certificate checks will be performed for Endpoint

Default Value: `false`

***

### .spec.s3.endpoint

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/ml/v1alpha1/storage_s3_spec.go#L30)</sup>

Endpoint specifies the S3 API-compatible endpoint which implements storage
Required

***

### .spec.s3.region

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/ml/v1alpha1/storage_s3_spec.go#L36)</sup>

Region defines the availability zone name. If empty, defaults to 'us-east-1'

Default Value: `""`

## Status

