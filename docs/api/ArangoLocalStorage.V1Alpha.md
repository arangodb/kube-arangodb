# API Reference for ArangoLocalStorage V1Alpha

## Spec

### .spec.localPath

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/storage/v1alpha/local_storage_spec.go#L36)</sup>

LocalPath setting specifies one or more local directories (on the nodes) used to create persistent volumes in.

***

### .spec.nodeSelector

Type: `object` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/storage/v1alpha/local_storage_spec.go#L43)</sup>

NodeSelector setting specifies which nodes the operator will provision persistent volumes on.

***

### .spec.podCustomization.priority

Type: `integer` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/storage/v1alpha/local_storage_pod_customization.go#L25)</sup>

Priority if defined, sets the priority for pods of storage provisioner

***

### .spec.privileged

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/storage/v1alpha/local_storage_spec.go#L45)</sup>

Privileged if set, passes Privileged flag to SecurityContext for pods of storage provisioner

***

### .spec.storageClass.isDefault

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/storage/v1alpha/storage_class_spec.go#L42)</sup>

IsDefault setting specifies if the created `StorageClass` will
be marked as default storage class.

Default Value: `false`

***

### .spec.storageClass.name

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/storage/v1alpha/storage_class_spec.go#L38)</sup>

Name setting specifies the name of the storage class that
created `PersistentVolume` will use.
If empty, this field defaults to the name of the `ArangoLocalStorage` object.
If a `StorageClass` with given name does not yet exist, it will be created.

Default Value: `""`

***

### .spec.storageClass.reclaimPolicy

Type: `core.PersistentVolumeReclaimPolicy` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/storage/v1alpha/storage_class_spec.go#L46)</sup>

ReclaimPolicy defines what happens to a persistent volume when released from its claim.

Links:
* [Documentation of core.PersistentVolumeReclaimPolicy](https://kubernetes.io/docs/concepts/storage/persistent-volumes#reclaiming)

***

### .spec.tolerations

Type: `[]core.Toleration` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/storage/v1alpha/local_storage_spec.go#L41)</sup>

Tolerations specifies the tolerations added to pods of storage provisioner

Links:
* [Documentation of core.Toleration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#toleration-v1-core)

