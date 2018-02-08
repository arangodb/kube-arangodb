# Storage

An ArangoDB cluster relies heavily on fast persistent storage.
The ArangoDB operator uses `PersistenVolumeClaims` to deliver
the storage to Pods that need them.

## Storage configuration

In the cluster resource, one can specify the type of storage
used by groups of servers using the `spec.<group>.storageClassName`
setting.

The amount of storage needed is configured using the
`spec.<group>.resources.requests.storage` setting.

Note that configuring storage is done per group of servers.
It is not possible to configure storage per individual
server.
