# ArangoLocalStorage Custom Resource

[Full CustomResourceDefinition reference ->](./api/ArangoLocalStorage.V1Alpha.md)

The ArangoDB Storage Operator creates and maintains ArangoDB
storage resources in a Kubernetes cluster, given a storage specification.
This storage specification is a `CustomResource` following
a `CustomResourceDefinition` created by the operator. It is not enabled by
default in the operator.

Example minimal storage definition:

```yaml
apiVersion: "storage.arangodb.com/v1alpha"
kind: "ArangoLocalStorage"
metadata:
  name: "example-arangodb-storage"
spec:
  storageClass:
    name: my-local-ssd
  localPath:
  - /mnt/big-ssd-disk
```

This definition results in:

- a `StorageClass` called `my-local-ssd`
- the dynamic provisioning of PersistentVolume's with
  a local volume on a node where the local volume starts
  in a sub-directory of `/mnt/big-ssd-disk`.
- the dynamic cleanup of PersistentVolume's (created by
  the operator) after one is released.

The provisioned volumes will have a capacity that matches
the requested capacity of volume claims.
