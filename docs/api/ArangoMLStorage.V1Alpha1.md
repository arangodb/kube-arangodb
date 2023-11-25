# API Reference for ArangoMLStorage V1Alpha1

## Spec

### .spec.backend.s3.allowInsecure

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/ml/v1alpha1/storage_spec_backend_s3.go#L39)</sup>

AllowInsecure if set to true, the Endpoint certificates won't be checked

Default Value: `false`

***

### .spec.backend.s3.bucketName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/ml/v1alpha1/storage_spec_backend_s3.go#L33)</sup>

BucketName specifies the name of the bucket
Required

***

### .spec.backend.s3.caSecretName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/ml/v1alpha1/storage_spec_backend_s3.go#L45)</sup>

CASecretName if not empty, the given secret will be used to check the authenticity of Endpoint
The specified `Secret`, must contain the following data fields:
- `ca.crt` PEM encoded public key of the CA certificate
- `ca.key` PEM encoded private key of the CA certificate

Default Value: `""`

***

### .spec.backend.s3.credentialsSecretName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/ml/v1alpha1/storage_spec_backend_s3.go#L36)</sup>

CredentialsSecretName specifies the name of the secret containing AccessKey and SecretKey for S3 API authorization
Required

***

### .spec.backend.s3.endpoint

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/ml/v1alpha1/storage_spec_backend_s3.go#L30)</sup>

Endpoint specifies the S3 API-compatible endpoint which implements storage
Required

***

### .spec.backend.s3.region

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/ml/v1alpha1/storage_spec_backend_s3.go#L48)</sup>

Region defines the availability zone name. If empty, defaults to 'us-east-1'

Default Value: `""`

***

### .spec.mode.sidecar.listenPort

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/ml/v1alpha1/storage_spec_mode_sidecar.go#L33)</sup>

ListenPort defines on which port the sidecar container will be listening for connections

Default Value: `9201`

***

### .spec.mode.sidecar.resources

Type: `core.ResourceRequirements` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/ml/v1alpha1/storage_spec_mode_sidecar.go#L38)</sup>

Resources holds resource requests & limits for container running the S3 proxy

Links:
* [Documentation of core.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#resourcerequirements-v1-core)

## Status

### .status.conditions

Type: `api.Conditions` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/ml/v1alpha1/storage_status.go#L28)</sup>

Conditions specific to the entire storage

