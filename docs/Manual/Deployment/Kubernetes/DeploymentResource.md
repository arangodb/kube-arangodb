# ArangoDeployment Custom Resource

The ArangoDB Deployment Operator creates and maintains ArangoDB deployments
in a Kubernetes cluster, given a deployment specification.
This deployment specification is a `CustomResource` following
a `CustomResourceDefinition` created by the operator.

Example minimal deployment definition of an ArangoDB database cluster:

```yaml
apiVersion: "database.arangodb.com/v1alpha"
kind: "ArangoDeployment"
metadata:
  name: "example-arangodb-cluster"
spec:
  mode: Cluster
```

Example more elaborate deployment definition:

```yaml
apiVersion: "database.arangodb.com/v1alpha"
kind: "ArangoDeployment"
metadata:
  name: "example-arangodb-cluster"
spec:
  mode: Cluster
  environment: Production
  agents:
    count: 3
    args:
      - --log.level=debug
    resources:
      requests:
        storage: 8Gi
    storageClassName: ssd
  dbservers:
    count: 5
    resources:
      requests:
        storage: 80Gi
    storageClassName: ssd
  coordinators:
    count: 3
  image: "arangodb/arangodb:3.3.4"
```

## Specification reference

Below you'll find all settings of the `ArangoDeployment` custom resource.
Several settings are for various groups of servers. These are indicated
with `<group>` where `<group>` can be any of:

- `agents` for all agents of a `Cluster` or `ActiveFailover` pair.
- `dbservers` for all dbservers of a `Cluster`.
- `coordinators` for all coordinators of a `Cluster`.
- `single` for all single servers of a `Single` instance or `ActiveFailover` pair.
- `syncmasters` for all syncmasters of a `Cluster`.
- `syncworkers` for all syncworkers of a `Cluster`.

### `spec.mode: string`

This setting specifies the type of deployment you want to create.
Possible values are:

- `Cluster` (default) Full cluster. Defaults to 3 agents, 3 dbservers & 3 coordinators.
- `ActiveFailover` Active-failover single pair. Defaults to 3 agents and 2 single servers.
- `Single` Single server only (note this does not provide high availability or reliability).

This setting cannot be changed after the deployment has been created.

### `spec.environment: string`

This setting specifies the type of environment in which the deployment is created.
Possible values are:

- `Development` (default) This value optimizes the deployment for development
  use. It is possible to run a deployment on a small number of nodes (e.g. minikube).
- `Production` This value optimizes the deployment for production use.
  It puts required affinity constraints on all pods to avoid agents & dbservers
  from running on the same machine.

### `spec.image: string`

This setting specifies the docker image to use for all ArangoDB servers.
In a `development` environment this setting defaults to `arangodb/arangodb:latest`.
For `production` environments this is a required setting without a default value.
It is highly recommend to use explicit version (not `latest`) for production
environments.

### `spec.imagePullPolicy: string`

This setting specifies the pull policy for the docker image to use for all ArangoDB servers.
Possible values are:

- `IfNotPresent` (default) to pull only when the image is not found on the node.
- `Always` to always pull the image before using it.

### `spec.imagePullSecrets: []string`

This setting specifies the list of image pull secrets for the docker image to use for all ArangoDB servers.

### `spec.storageEngine: string`

This setting specifies the type of storage engine used for all servers
in the cluster.
Possible values are:

- `MMFiles` To use the MMFiles storage engine.
- `RocksDB` (default) To use the RocksDB storage engine.

This setting cannot be changed after the cluster has been created.

### `spec.downtimeAllowed: bool`

This setting is used to allow automatic reconciliation actions that yield
some downtime of the ArangoDB deployment.
When this setting is set to `false` (the default), no automatic action that
may result in downtime is allowed.
If the need for such an action is detected, an event is added to the `ArangoDeployment`.

Once this setting is set to `true`, the automatic action is executed.

Operations that may result in downtime are:

- Rotating TLS CA certificate

Note: It is still possible that there is some downtime when the Kubernetes
cluster is down, or in a bad state, irrespective of the value of this setting.

### `spec.rocksdb.encryption.keySecretName`

This setting specifies the name of a Kubernetes `Secret` that contains
an encryption key used for encrypting all data stored by ArangoDB servers.
When an encryption key is used, encryption of the data in the cluster is enabled,
without it encryption is disabled.
The default value is empty.

This requires the Enterprise version.

The encryption key cannot be changed after the cluster has been created.

The secret specified by this setting, must have a data field named 'key' containing
an encryption key that is exactly 32 bytes long.

### `spec.networkAttachedVolumes: bool`

The default of this option is `false`. If set to `true`, and the
deployed ArangoDB version is new enough (>= 3.4.8 for 3.4 and >= 3.5.1
for 3.5), a `ResignLeaderShip` operation
will be triggered when a dbserver pod is evicted (rather than a
`CleanOutServer` operation). Furthermore, the pod will simply be
redeployed on a different node, rather than cleaned and retired and
replaced by a new member. You must only set this option to `true` if
your persistent volumes are "movable" in the sense that they can be
mounted from a different k8s node, like in the case of network attached
volumes. If your persistent volumes are tied to a specific pod, you
must leave this option on `false`.

### `spec.externalAccess.type: string`

This setting specifies the type of `Service` that will be created to provide
access to the ArangoDB deployment from outside the Kubernetes cluster.
Possible values are:

- `None` To limit access to application running inside the Kubernetes cluster.
- `LoadBalancer` To create a `Service` of type `LoadBalancer` for the ArangoDB deployment.
- `NodePort` To create a `Service` of type `NodePort` for the ArangoDB deployment.
- `Auto` (default) To create a `Service` of type `LoadBalancer` and fallback to a `Service` or type `NodePort` when the
  `LoadBalancer` is not assigned an IP address.

### `spec.externalAccess.loadBalancerIP: string`

This setting specifies the IP used to for the LoadBalancer to expose the ArangoDB deployment on.
This setting is used when `spec.externalAccess.type` is set to `LoadBalancer` or `Auto`.

If you do not specify this setting, an IP will be chosen automatically by the load-balancer provisioner.

### `spec.externalAccess.loadBalancerSourceRanges: []string`

If specified and supported by the platform (cloud provider), this will restrict traffic through the cloud-provider
load-balancer will be restricted to the specified client IPs. This field will be ignored if the
cloud-provider does not support the feature.

More info: https://kubernetes.io/docs/tasks/access-application-cluster/configure-cloud-provider-firewall/

### `spec.externalAccess.nodePort: int`

This setting specifies the port used to expose the ArangoDB deployment on.
This setting is used when `spec.externalAccess.type` is set to `NodePort` or `Auto`.

If you do not specify this setting, a random port will be chosen automatically.

### `spec.externalAccess.advertisedEndpoint: string`

This setting specifies the advertised endpoint for all coordinators.

### `spec.auth.jwtSecretName: string`

This setting specifies the name of a kubernetes `Secret` that contains
the JWT token used for accessing all ArangoDB servers.
When no name is specified, it defaults to `<deployment-name>-jwt`.
To disable authentication, set this value to `None`.

If you specify a name of a `Secret`, that secret must have the token
in a data field named `token`.

If you specify a name of a `Secret` that does not exist, a random token is created
and stored in a `Secret` with given name.

Changing a JWT token results in stopping the entire cluster
and restarting it.

### `spec.tls.caSecretName: string`

This setting specifies the name of a kubernetes `Secret` that contains
a standard CA certificate + private key used to sign certificates for individual
ArangoDB servers.
When no name is specified, it defaults to `<deployment-name>-ca`.
To disable authentication, set this value to `None`.

If you specify a name of a `Secret` that does not exist, a self-signed CA certificate + key is created
and stored in a `Secret` with given name.

The specified `Secret`, must contain the following data fields:

- `ca.crt` PEM encoded public key of the CA certificate
- `ca.key` PEM encoded private key of the CA certificate

### `spec.tls.altNames: []string`

This setting specifies a list of alternate names that will be added to all generated
certificates. These names can be DNS names or email addresses.
The default value is empty.

### `spec.tls.ttl: duration`

This setting specifies the time to live of all generated
server certificates.
The default value is `2160h` (about 3 month).

When the server certificate is about to expire, it will be automatically replaced
by a new one and the affected server will be restarted.

Note: The time to live of the CA certificate (when created automatically)
will be set to 10 years.

### `spec.sync.enabled: bool`

This setting enables/disables support for data center 2 data center
replication in the cluster. When enabled, the cluster will contain
a number of `syncmaster` & `syncworker` servers.
The default value is `false`.

### `spec.sync.externalAccess.type: string`

This setting specifies the type of `Service` that will be created to provide
access to the ArangoSync syncMasters from outside the Kubernetes cluster.
Possible values are:

- `None` To limit access to applications running inside the Kubernetes cluster.
- `LoadBalancer` To create a `Service` of type `LoadBalancer` for the ArangoSync SyncMasters.
- `NodePort` To create a `Service` of type `NodePort` for the ArangoSync SyncMasters.
- `Auto` (default) To create a `Service` of type `LoadBalancer` and fallback to a `Service` or type `NodePort` when the
  `LoadBalancer` is not assigned an IP address.

Note that when you specify a value of `None`, a `Service` will still be created, but of type `ClusterIP`.

### `spec.sync.externalAccess.loadBalancerIP: string`

This setting specifies the IP used for the LoadBalancer to expose the ArangoSync SyncMasters on.
This setting is used when `spec.sync.externalAccess.type` is set to `LoadBalancer` or `Auto`.

If you do not specify this setting, an IP will be chosen automatically by the load-balancer provisioner.

### `spec.sync.externalAccess.nodePort: int`

This setting specifies the port used to expose the ArangoSync SyncMasters on.
This setting is used when `spec.sync.externalAccess.type` is set to `NodePort` or `Auto`.

If you do not specify this setting, a random port will be chosen automatically.

### `spec.sync.externalAccess.loadBalancerSourceRanges: []string`

If specified and supported by the platform (cloud provider), this will restrict traffic through the cloud-provider
load-balancer will be restricted to the specified client IPs. This field will be ignored if the
cloud-provider does not support the feature.

More info: https://kubernetes.io/docs/tasks/access-application-cluster/configure-cloud-provider-firewall/


### `spec.sync.externalAccess.masterEndpoint: []string`

This setting specifies the master endpoint(s) advertised by the ArangoSync SyncMasters.
If not set, this setting defaults to:

- If `spec.sync.externalAccess.loadBalancerIP` is set, it defaults to `https://<load-balancer-ip>:<8629>`.
- Otherwise it defaults to `https://<sync-service-dns-name>:<8629>`.

### `spec.sync.externalAccess.accessPackageSecretNames: []string`

This setting specifies the names of zero of more `Secrets` that will be created by the deployment
operator containing "access packages". An access package contains those `Secrets` that are needed
to access the SyncMasters of this `ArangoDeployment`.

By removing a name from this setting, the corresponding `Secret` is also deleted.
Note that to remove all access packages, leave an empty array in place (`[]`).
Completely removing the setting results in not modifying the list.

See [the `ArangoDeploymentReplication` specification](./DeploymentReplicationResource.md) for more information
on access packages.

### `spec.sync.auth.jwtSecretName: string`

This setting specifies the name of a kubernetes `Secret` that contains
the JWT token used for accessing all ArangoSync master servers.
When not specified, the `spec.auth.jwtSecretName` value is used.

If you specify a name of a `Secret` that does not exist, a random token is created
and stored in a `Secret` with given name.

### `spec.sync.auth.clientCASecretName: string`

This setting specifies the name of a kubernetes `Secret` that contains
a PEM encoded CA certificate used for client certificate verification
in all ArangoSync master servers.
This is a required setting when `spec.sync.enabled` is `true`.
The default value is empty.

### `spec.sync.mq.type: string`

This setting sets the type of message queue used by ArangoSync.
Possible values are:

- `Direct` (default) for direct HTTP connections between the 2 data centers.

### `spec.sync.tls.caSecretName: string`

This setting specifies the name of a kubernetes `Secret` that contains
a standard CA certificate + private key used to sign certificates for individual
ArangoSync master servers.

When no name is specified, it defaults to `<deployment-name>-sync-ca`.

If you specify a name of a `Secret` that does not exist, a self-signed CA certificate + key is created
and stored in a `Secret` with given name.

The specified `Secret`, must contain the following data fields:

- `ca.crt` PEM encoded public key of the CA certificate
- `ca.key` PEM encoded private key of the CA certificate

### `spec.sync.tls.altNames: []string`

This setting specifies a list of alternate names that will be added to all generated
certificates. These names can be DNS names or email addresses.
The default value is empty.

### `spec.sync.monitoring.tokenSecretName: string`

This setting specifies the name of a kubernetes `Secret` that contains
the bearer token used for accessing all monitoring endpoints of all ArangoSync
servers.
When not specified, no monitoring token is used.
The default value is empty.

### `spec.disableIPv6: bool`

This setting prevents the use of IPv6 addresses by ArangoDB servers.
The default is `false`. 

This setting cannot be changed after the deployment has been created.

### `spec.restoreFrom: string`

This setting specifies a `ArangoBackup` resource name the cluster should be restored from.

After a restore or failure to do so, the status of the deployment contains information about the
restore operation in the `restore` key.

It will contain some of the following fields:
- _requestedFrom_: name of the `ArangoBackup` used to restore from.
- _message_: optional message explaining why the restore failed.
- _state_: state indicating if the restore was successful or not. Possilbe values: `Restoring`, `Restored`, `RestoreFailed`

If the `restoreFrom` key is removed from the spec, the `restore` key is deleted as well.

A new restore attempt is made if and only if either in the status restore is not set or if spec.restoreFrom and status.requestedFrom are different.

### `spec.license.secretName: string`

This setting specifies the name of a kubernetes `Secret` that contains
the license key token used for enterprise images. This value is not used for
the community edition.

### `spec.bootstrap.passwordSecretNames.root: string`

This setting specifies a secret name for the credentials of the root user.

When a deployment is created the operator will setup the root user account 
according to the credentials given by the secret. If the secret doesn't exist
the operator creates a secret with a random password.

There are two magic values for the secret name:
- `None` specifies no action. This disables root password randomization. This is the default value. (Thus the root password is empty - not recommended)
- `Auto` specifies automatic name generation, which is `<deploymentname>-root-password`. 

### `spec.metrics.enabled: bool`

If this is set to `true`, the operator runs a sidecar container for
every DBserver pod and every coordinator pod. The sidecar container runs
the ArangoDB-exporter and exposes metrics of the corresponding `arangod`
instance in Prometheus format on port 9101 under path `/metrics`. You
also have to specify a string for `spec.metrics.image`, which is the
Docker image name of the `arangodb-exporter`. At the time of this
writing you should use `arangodb/arangodb-exporter:0.1.6`. See [this
repository](https://github.com/arangodb-helper/arangodb-exporter) for
the latest version. If the image name is left empty, the same image as
for the main deployment is used. Note however, that current ArangoDB
releases (<= 3.4.5) do not ship the exporter in their image. This is
going to change in the future.

In addition to the sidecar containers the operator will deploy a service
to access the exporter ports (from within the k8s cluster), and a
resource of type `ServiceMonitor`, provided the corresponding custom
resource definition is deployed in the k8s cluster. If you are running
Prometheus in the same k8s cluster with the Prometheus operator, this
will be the case. The `ServiceMonitor` will have the following labels
set:

  - `app: arangodb`
  - `arango_deployment: YOUR_DEPLOYMENT_NAME`
  - `context: metrics`
  - `metrics: prometheus`

This makes it possible that you configure your Prometheus deployment to
automatically start monitoring on the available Prometheus feeds. To
this end, you must configure the `serviceMonitorSelector` in the specs
of your Prometheus deployment to match these labels. For example:

```yaml
  serviceMonitorSelector:
    matchLabels:
      metrics: prometheus
```

would automatically select all pods of all ArangoDB cluster deployments
which have metrics enabled.

### `spec.metrics.image: string`

See above, this is the name of the Docker image for the ArangoDB
exporter to expose metrics. If empty, the same image as for the main
deployment is used.

### `spec.<group>.count: number`

This setting specifies the number of servers to start for the given group.
For the agent group, this value must be a positive, odd number.
The default value is `3` for all groups except `single` (there the default is `1`
for `spec.mode: Single` and `2` for `spec.mode: ActiveFailover`).

For the `syncworkers` group, it is highly recommended to use the same number
as for the `dbservers` group.

### `spec.<group>.minCount: number`

Specifies a minimum for the count of servers. If set, a specification is invalid if `count < minCount`.

### `spec.<group>.maxCount: number`

Specifies a maximum for the count of servers. If set, a specification is invalid if `count > maxCount`.

### `spec.<group>.args: []string`

This setting specifies additional commandline arguments passed to all servers of this group.
The default value is an empty array.

### `spec.<group>.resources: ResourceRequirements`

This setting specifies the resources required by pods of this group. This includes requests and limits.

See https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/ for details.

### `spec.<group>.volumeClaimTemplate.Spec: PersistentVolumeClaimSpec`

Specifies a volumeClaimTemplate used by operator to create to volume claims for pods of this group.
This setting is not available for group `coordinators`, `syncmasters` & `syncworkers`.

The default value describes a volume with `8Gi` storage, `ReadWriteOnce` access mode and volume mode set to `PersistentVolumeFilesystem`. 

If this field is not set and `spec.<group>.resources.requests.storage` is set, then a default volume claim
with size as specified by `spec.<group>.resources.requests.storage` will be created. In that case `storage`
and `iops` is not forwarded to the pods resource requirements.

### `spec.<group>.serviceAccountName: string`

This setting specifies the `serviceAccountName` for the `Pods` created
for each server of this group. If empty, it defaults to using the
`default` service account.

Using an alternative `ServiceAccount` is typically used to separate access rights.
The ArangoDB deployments need some very minimal access rights. With the
deployment of the operator, we grant the following rights for the `default`
service account:

```
rules:
- apiGroups:
  - ""
  resources:
  - pods
  verbs:
  - get
```

If you are using a different service account, please grant these rights
to that service account.

### `spec.<group>.priorityClassName: string`

Priority class name for pods of this group. Will be forwarded to the pod spec. [Kubernetes documentation](https://kubernetes.io/docs/concepts/configuration/pod-priority-preemption/)

### `spec.<group>.probes.livenessProbeDisabled: bool`

If set to true, the operator does not generate a liveness probe for new pods belonging to this group.

### `spec.<group>.probes.readinessProbeDisabled: bool`

If set to true, the operator does not generate a readiness probe for new pods belonging to this group.

### `spec.<group>.tolerations: []Toleration`

This setting specifies the `tolerations` for the `Pod`s created
for each server of this group.

By default, suitable tolerations are set for the following keys with the `NoExecute` effect:

- `node.kubernetes.io/not-ready`
- `node.kubernetes.io/unreachable`
- `node.alpha.kubernetes.io/unreachable` (will be removed in future version)

For more information on tolerations, consult the [Kubernetes documentation](https://kubernetes.io/docs/concepts/configuration/taint-and-toleration/).

### `spec.<group>.nodeSelector: map[string]string`

This setting specifies a set of labels to be used as `nodeSelector` for Pods of this node.

For more information on node selectors, consult the [Kubernetes documentation](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/).

## Deprecated Fields

### `spec.<group>.resources.requests.storage: storageUnit`

This setting specifies the amount of storage required for each server of this group.
The default value is `8Gi`.

This setting is not available for group `coordinators`, `syncmasters` & `syncworkers`
because servers in these groups do not need persistent storage.

Please use VolumeClaimTemplate from now on. This field is not considered if 
VolumeClaimTemplate is set. Note however, that the information in requests 
is completely handed over to the pod in this case.

### `spec.<group>.storageClassName: string`

This setting specifies the `storageClass` for the `PersistentVolume`s created
for each server of this group.

This setting is not available for group `coordinators`, `syncmasters` & `syncworkers`
because servers in these groups do not need persistent storage.

Please use VolumeClaimTemplate from now on. This field is not considered if 
VolumeClaimTemplate is set. Note however, that the information in requests 
is completely handed over to the pod in this case.
