# Storage

An ArangoDB cluster relies heavily on fast persistent storage.
The ArangoDB operator uses PersistenVolumeClaim's to deliver
this storages to Pods that need them.

TODO how to specify storage class (and other parameters)

Q: Do we want volumes other than PersistentVolumeClaims?
