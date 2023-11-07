# API Reference for ArangoLocalStorage V1Alpha

## Spec

### .spec.localPath: array

LocalPath setting specifies one or more local directories (on the nodes) used to create persistent volumes in.

[Code Reference](/pkg/apis/storage/v1alpha/local_storage_spec.go#L36)

### .spec.nodeSelector: object

NodeSelector setting specifies which nodes the operator will provision persistent volumes on.

[Code Reference](/pkg/apis/storage/v1alpha/local_storage_spec.go#L43)

### .spec.podCustomization.priority: integer

Priority if defined, sets the priority for pods of storage provisioner

[Code Reference](/pkg/apis/storage/v1alpha/local_storage_pod_customization.go#L25)

### .spec.privileged: boolean

Privileged if set, passes Privileged flag to SecurityContext for pods of storage provisioner

[Code Reference](/pkg/apis/storage/v1alpha/local_storage_spec.go#L45)

### .spec.storageClass.isDefault: boolean

IsDefault setting specifies if the created `StorageClass` will
be marked as default storage class.

Default Value: false

[Code Reference](/pkg/apis/storage/v1alpha/storage_class_spec.go#L42)

### .spec.storageClass.name: string

Name setting specifies the name of the storage class that
created `PersistentVolume` will use.
If empty, this field defaults to the name of the `ArangoLocalStorage` object.
If a `StorageClass` with given name does not yet exist, it will be created.

Default Value: ""

[Code Reference](/pkg/apis/storage/v1alpha/storage_class_spec.go#L38)

### .spec.storageClass.reclaimPolicy: core.PersistentVolumeReclaimPolicy

ReclaimPolicy defines what happens to a persistent volume when released from its claim.

Links:
* [Documentation of core.PersistentVolumeReclaimPolicy](https://kubernetes.io/docs/concepts/storage/persistent-volumes#reclaiming)

[Code Reference](/pkg/apis/storage/v1alpha/storage_class_spec.go#L46)

### .spec.tolerations: []core.Toleration

Tolerations specifies the tolerations added to pods of storage provisioner

Links:
* [Documentation of core.Toleration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#toleration-v1-core)

[Code Reference](/pkg/apis/storage/v1alpha/local_storage_spec.go#L41)

