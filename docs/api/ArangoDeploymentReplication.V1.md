# API Reference for ArangoDeploymentReplication V1

## Spec

### .spec.cancellation.ensureInSync

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/replication/v1/replication_spec.go#L38)</sup>

EnsureInSync if it is true then during cancellation process data consistency is required.
Default value is true.

***

### .spec.cancellation.sourceReadOnly

Type: `boolean` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/replication/v1/replication_spec.go#L41)</sup>

SourceReadOnly if it true then after cancellation source data center should be in read-only mode.
Default value is false.

***

### .spec.destination.auth.keyfileSecretName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/replication/v1/endpoint_authentication_spec.go#L37)</sup>

KeyfileSecretName holds the name of a Secret containing a client authentication
certificate formatted at keyfile in a `tls.keyfile` field.
If `userSecretName` has not been set,
the client authentication certificate found in the secret with this name is also used to configure
the synchronization and fetch the synchronization status.

***

### .spec.destination.auth.userSecretName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/replication/v1/endpoint_authentication_spec.go#L42)</sup>

UserSecretName holds the name of a Secret containing a `username` & `password`
field used for basic authentication.
The user identified by the username must have write access in the `_system` database
of the ArangoDB cluster at the endpoint.

***

### .spec.destination.deploymentName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/replication/v1/endpoint_spec.go#L36)</sup>

DeploymentName holds the name of an ArangoDeployment resource.
If set, this provides default values for masterEndpoint, auth & tls.

***

### .spec.destination.masterEndpoint

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/replication/v1/endpoint_spec.go#L42)</sup>

MasterEndpoint holds a list of URLs used to reach the syncmaster(s)
Use this setting if the source cluster is not running inside a Kubernetes cluster
that is reachable from the Kubernetes cluster the `ArangoDeploymentReplication` resource is deployed in.
Specifying this setting and `deploymentName` at the same time is not allowed.

Default Value: `[]`

***

### .spec.destination.tls.caSecretName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/replication/v1/endpoint_tls_spec.go#L34)</sup>

CASecretName holds the name of a Secret containing a ca.crt public key for TLS validation.
This setting is required, unless `deploymentName` has been set.

***

### .spec.source.auth.keyfileSecretName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/replication/v1/endpoint_authentication_spec.go#L37)</sup>

KeyfileSecretName holds the name of a Secret containing a client authentication
certificate formatted at keyfile in a `tls.keyfile` field.
If `userSecretName` has not been set,
the client authentication certificate found in the secret with this name is also used to configure
the synchronization and fetch the synchronization status.

***

### .spec.source.auth.userSecretName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/replication/v1/endpoint_authentication_spec.go#L42)</sup>

UserSecretName holds the name of a Secret containing a `username` & `password`
field used for basic authentication.
The user identified by the username must have write access in the `_system` database
of the ArangoDB cluster at the endpoint.

***

### .spec.source.deploymentName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/replication/v1/endpoint_spec.go#L36)</sup>

DeploymentName holds the name of an ArangoDeployment resource.
If set, this provides default values for masterEndpoint, auth & tls.

***

### .spec.source.masterEndpoint

Type: `array` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/replication/v1/endpoint_spec.go#L42)</sup>

MasterEndpoint holds a list of URLs used to reach the syncmaster(s)
Use this setting if the source cluster is not running inside a Kubernetes cluster
that is reachable from the Kubernetes cluster the `ArangoDeploymentReplication` resource is deployed in.
Specifying this setting and `deploymentName` at the same time is not allowed.

Default Value: `[]`

***

### .spec.source.tls.caSecretName

Type: `string` <sup>[\[ref\]](https://github.com/arangodb/kube-arangodb/blob/1.2.35/pkg/apis/replication/v1/endpoint_tls_spec.go#L34)</sup>

CASecretName holds the name of a Secret containing a ca.crt public key for TLS validation.
This setting is required, unless `deploymentName` has been set.

