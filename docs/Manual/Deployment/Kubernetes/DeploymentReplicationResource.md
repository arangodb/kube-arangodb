# ArangoDeploymentReplication Custom Resource

The ArangoDB Replication Operator creates and maintains ArangoDB
`arangosync` configurations in a Kubernetes cluster, given a replication specification.
This replication specification is a `CustomResource` following
a `CustomResourceDefinition` created by the operator.

Example minimal replication definition:

```yaml
apiVersion: "database.arangodb.com/v1alpha"
kind: "ArangoDeploymentReplication"
metadata:
  name: "replication-from-a-to-b"
spec:
  source:
    deploymentName: cluster-a
  destination:
    deploymentName: cluster-b
  auth:
    clientSecretName: client-auth-cert
```

This definition results in:

- the arangosync `SyncMaster` in deployment `cluster-b` is called to configure a synchronization
  from `cluster-a` to `cluster-b`, using the client authentication certificate stored in
  `Secret` `client-auth-cert`.

## Specification reference

Below you'll find all settings of the `ArangoDeploymentReplication` custom resource.

### `spec.source.deploymentName: string`

This setting specifies the name of an `ArangoDeployment` resource that runs a cluster
with sync enabled.

This cluster configured as the replication source.

### `spec.source.deploymentNamespace: string`

This setting specifies the Kubernetes namespace of an `ArangoDeployment` resource specified in `spec.source.deploymentName`.

If this setting is empty, the namespace of the `ArangoDeploymentReplication` is used.

### `spec.source.masterEndpoints: []string`

This setting specifies zero or more master endpoint URL's of the source cluster.

Use this setting if the source cluster is not running inside a Kubernetes cluster
that is reachable from the Kubernetes cluster the `ArangoDeploymentReplication` resource is deployed in.

Specifying this setting and `spec.source.deploymentName` at the same time is not allowed.

### `spec.destination.deploymentName: string`

This setting specifies the name of an `ArangoDeployment` resource that runs a cluster
with sync enabled.

This cluster configured as the replication destination.

### `spec.destination.deploymentNamespace: string`

This setting specifies the Kubernetes namespace of an `ArangoDeployment` resource specified in `spec.destination.deploymentName`.

If this setting is empty, the namespace of the `ArangoDeploymentReplication` is used.

### `spec.destination.masterEndpoints: []string`

This setting specifies zero or more master endpoint URL's of the destination cluster.

Use this setting if the destination cluster is not running inside a Kubernetes cluster
that is reachable from the Kubernetes cluster the `ArangoDeploymentReplication` resource is deployed in.

Specifying this setting and `spec.destination.deploymentName` at the same time is not allowed.

### `spec.auth.clientSecretName: string`

This setting specifies the name of a `Secret` containing a client authentication certificate,
used to authenticate the SyncMaster in the destination cluster with the SyncMaster in the
source cluster.
