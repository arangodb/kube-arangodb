# API Reference for ArangoMLExtension V1Alpha1

## Spec

### .spec.image

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/image.go#L31)</sup>

Image define image details

***

### .spec.init.image

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/image.go#L31)</sup>

Image define image details

***

### .spec.init.pullPolicy

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/image.go#L35)</sup>

PullPolicy define Image pull policy

Default Value: `IfNotPresent`

***

### .spec.init.pullSecrets

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/image.go#L38)</sup>

PullSecrets define Secrets used to pull Image from registry

***

### .spec.metadataService.local.arangoMLFeatureStore

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/ml/v1alpha1/extension_spec_metadata_service.go#L65)</sup>

ArangoMLFeatureStoreDatabase define Database name to be used as MetadataService Backend in ArangoMLFeatureStoreDatabase

Default Value: `arangomlfeaturestore`

***

### .spec.metadataService.local.arangoPipeDatabase

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/ml/v1alpha1/extension_spec_metadata_service.go#L61)</sup>

ArangoPipeDatabase define Database name to be used as MetadataService Backend in ArangoPipe

Default Value: `arangopipe`

***

### .spec.pullPolicy

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/image.go#L35)</sup>

PullPolicy define Image pull policy

Default Value: `IfNotPresent`

***

### .spec.pullSecrets

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/image.go#L38)</sup>

PullSecrets define Secrets used to pull Image from registry

***

### .spec.storage.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/object.go#L46)</sup>

Name of the object

***

### .spec.storage.namespace

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/object.go#L49)</sup>

Namespace of the object. Should default to the namespace of the parent object

***

### .spec.storage.uid

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/object.go#L52)</sup>

UID keeps the information about object UID

## Status

### .status.conditions

Type: `api.Conditions` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/ml/v1alpha1/extension_status.go#L31)</sup>

Conditions specific to the entire extension

***

### .status.metadataService.local.arangoMLFeatureStore

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/ml/v1alpha1/extension_status_metadata_service.go#L38)</sup>

ArangoMLFeatureStoreDatabase define Database name to be used as MetadataService Backend

***

### .status.metadataService.local.arangoPipe

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/ml/v1alpha1/extension_status_metadata_service.go#L35)</sup>

ArangoPipeDatabase define Database name to be used as MetadataService Backend

***

### .status.metadataService.secret.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/object.go#L46)</sup>

Name of the object

***

### .status.metadataService.secret.namespace

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/object.go#L49)</sup>

Namespace of the object. Should default to the namespace of the parent object

***

### .status.metadataService.secret.uid

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/object.go#L52)</sup>

UID keeps the information about object UID

***

### .status.serviceAccount.cluster.binding.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/object.go#L46)</sup>

Name of the object

***

### .status.serviceAccount.cluster.binding.namespace

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/object.go#L49)</sup>

Namespace of the object. Should default to the namespace of the parent object

***

### .status.serviceAccount.cluster.binding.uid

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/object.go#L52)</sup>

UID keeps the information about object UID

***

### .status.serviceAccount.cluster.role.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/object.go#L46)</sup>

Name of the object

***

### .status.serviceAccount.cluster.role.namespace

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/object.go#L49)</sup>

Namespace of the object. Should default to the namespace of the parent object

***

### .status.serviceAccount.cluster.role.uid

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/object.go#L52)</sup>

UID keeps the information about object UID

***

### .status.serviceAccount.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/object.go#L46)</sup>

Name of the object

***

### .status.serviceAccount.namespace

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/object.go#L49)</sup>

Namespace of the object. Should default to the namespace of the parent object

***

### .status.serviceAccount.namespaced.binding.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/object.go#L46)</sup>

Name of the object

***

### .status.serviceAccount.namespaced.binding.namespace

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/object.go#L49)</sup>

Namespace of the object. Should default to the namespace of the parent object

***

### .status.serviceAccount.namespaced.binding.uid

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/object.go#L52)</sup>

UID keeps the information about object UID

***

### .status.serviceAccount.namespaced.role.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/object.go#L46)</sup>

Name of the object

***

### .status.serviceAccount.namespaced.role.namespace

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/object.go#L49)</sup>

Namespace of the object. Should default to the namespace of the parent object

***

### .status.serviceAccount.namespaced.role.uid

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/object.go#L52)</sup>

UID keeps the information about object UID

***

### .status.serviceAccount.uid

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/shared/v1/object.go#L52)</sup>

UID keeps the information about object UID

