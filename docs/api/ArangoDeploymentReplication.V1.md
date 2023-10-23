# API Reference for ArangoDeploymentReplication V1

## Spec

### .spec.cancellation.ensureInSync: bool

EnsureInSync if it is true then during cancellation process data consistency is required.
Default value is true.

[Code Reference](/pkg/apis/replication/v1/replication_spec.go#L38)

### .spec.cancellation.sourceReadOnly: bool

SourceReadOnly if it true then after cancellation source data center should be in read-only mode.
Default value is false.

[Code Reference](/pkg/apis/replication/v1/replication_spec.go#L41)

### .spec.destination.auth.keyfileSecretName: string

KeyfileSecretName holds the name of a Secret containing a client authentication
certificate formatted at keyfile in a `tls.keyfile` field.
If `userSecretName` has not been set,
the client authentication certificate found in the secret with this name is also used to configure
the synchronization and fetch the synchronization status.

[Code Reference](/pkg/apis/replication/v1/endpoint_authentication_spec.go#L37)

### .spec.destination.auth.userSecretName: string

UserSecretName holds the name of a Secret containing a `username` & `password`
field used for basic authentication.
The user identified by the username must have write access in the `_system` database
of the ArangoDB cluster at the endpoint.

[Code Reference](/pkg/apis/replication/v1/endpoint_authentication_spec.go#L42)

### .spec.destination.deploymentName: string

DeploymentName holds the name of an ArangoDeployment resource.
If set, this provides default values for masterEndpoint, auth & tls.

[Code Reference](/pkg/apis/replication/v1/endpoint_spec.go#L36)

### .spec.destination.masterEndpoint: []string

MasterEndpoint holds a list of URLs used to reach the syncmaster(s)
Use this setting if the source cluster is not running inside a Kubernetes cluster
that is reachable from the Kubernetes cluster the `ArangoDeploymentReplication` resource is deployed in.
Specifying this setting and `deploymentName` at the same time is not allowed.

Default Value: []

[Code Reference](/pkg/apis/replication/v1/endpoint_spec.go#L42)

### .spec.destination.tls.caSecretName: string

CASecretName holds the name of a Secret containing a ca.crt public key for TLS validation.
This setting is required, unless `deploymentName` has been set.

[Code Reference](/pkg/apis/replication/v1/endpoint_tls_spec.go#L34)

### .spec.source.auth.keyfileSecretName: string

KeyfileSecretName holds the name of a Secret containing a client authentication
certificate formatted at keyfile in a `tls.keyfile` field.
If `userSecretName` has not been set,
the client authentication certificate found in the secret with this name is also used to configure
the synchronization and fetch the synchronization status.

[Code Reference](/pkg/apis/replication/v1/endpoint_authentication_spec.go#L37)

### .spec.source.auth.userSecretName: string

UserSecretName holds the name of a Secret containing a `username` & `password`
field used for basic authentication.
The user identified by the username must have write access in the `_system` database
of the ArangoDB cluster at the endpoint.

[Code Reference](/pkg/apis/replication/v1/endpoint_authentication_spec.go#L42)

### .spec.source.deploymentName: string

DeploymentName holds the name of an ArangoDeployment resource.
If set, this provides default values for masterEndpoint, auth & tls.

[Code Reference](/pkg/apis/replication/v1/endpoint_spec.go#L36)

### .spec.source.masterEndpoint: []string

MasterEndpoint holds a list of URLs used to reach the syncmaster(s)
Use this setting if the source cluster is not running inside a Kubernetes cluster
that is reachable from the Kubernetes cluster the `ArangoDeploymentReplication` resource is deployed in.
Specifying this setting and `deploymentName` at the same time is not allowed.

Default Value: []

[Code Reference](/pkg/apis/replication/v1/endpoint_spec.go#L42)

### .spec.source.tls.caSecretName: string

CASecretName holds the name of a Secret containing a ca.crt public key for TLS validation.
This setting is required, unless `deploymentName` has been set.

[Code Reference](/pkg/apis/replication/v1/endpoint_tls_spec.go#L34)

