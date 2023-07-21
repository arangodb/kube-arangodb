# API Reference for ArangoDeployment V1

## Spec

### .spec.agents: ServerGroupSpec

Agents contains specification for Agency pods running in deployment mode `Cluster` or `ActiveFailover`.

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L246)

### .spec.allowUnsafeUpgrade: bool

AllowUnsafeUpgrade determines if upgrade on missing member or with not in sync shards is allowed

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L210)

### .spec.annotations: map[string]string

Annotations specifies the annotations added to all ArangoDeployment owned resources (pods, services, PVC’s, PDB’s).

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L179)

### .spec.annotationsIgnoreList: []string

AnnotationsIgnoreList list regexp or plain definitions which annotations should be ignored

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L181)

### .spec.annotationsMode: string

AnnotationsMode defines annotations mode which should be use while overriding annotations.
Possible values are:
- `disabled` disable annotations/labels override. Default if there is no annotations/labels set in ArangoDeployment
- `append` add new annotations/labels without affecting old ones
- `replace` replace existing annotations/labels

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L187)

### .spec.architecture: ArangoDeploymentArchitecture

Architecture defines the list of supported architectures.

Default Value: ['amd64']

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L294)

### .spec.auth: AuthenticationSpec

Authentication holds authentication configuration settings

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L219)

### .spec.bootstrap: BootstrapSpec

Bootstrap contains information for cluster bootstrapping

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L275)

### .spec.chaos: ChaosSpec

ChaosSpec can be used for chaos-monkey testing of your ArangoDeployment

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L269)

### .spec.ClusterDomain: string

ClusterDomain define domain used in the kubernetes cluster.
Required only of domain is not set to default (cluster.local)

Default Value: cluster.local

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L282)

### .spec.communicationMethod: string

CommunicationMethod define communication method used in deployment

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L284)

### .spec.coordinators: ServerGroupSpec

Coordinators contains specification for Coordinator pods running in deployment mode `Cluster` or `ActiveFailover`.

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L252)

### .spec.database: DatabaseSpec

Database holds information about database state, like maintenance mode

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L240)

### .spec.dbservers: ServerGroupSpec

DBServers contains specification for DBServer pods running in deployment mode `Cluster` or `ActiveFailover`.

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L249)

### .spec.disableIPv6: bool

DisableIPv6 setting prevents the use of IPv6 addresses by ArangoDB servers.
This setting cannot be changed after the deployment has been created.

Default Value: false

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L164)

### .spec.downtimeAllowed: bool

DowntimeAllowed setting is used to allow automatic reconciliation actions that yield some downtime of the ArangoDB deployment.
When this setting is set to false, no automatic action that may result in downtime is allowed.
If the need for such an action is detected, an event is added to the ArangoDeployment.
Once this setting is set to true, the automatic action is executed.
Operations that may result in downtime are:
- Rotating TLS CA certificate
Note: It is still possible that there is some downtime when the Kubernetes cluster is down, or in a bad state, irrespective of the value of this setting.

Default Value: false

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L160)

### .spec.environment: string

Environment setting specifies the type of environment in which the deployment is created.
Possible values are:
- `Development` This value optimizes the deployment for development use. It is possible to run a deployment on a small number of nodes (e.g. minikube).
- `Production` This value optimizes the deployment for production use. It puts required affinity constraints on all pods to avoid Agents & DB-Servers from running on the same machine.

Default Value: Development

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L127)

### .spec.externalAccess: ExternalAccessSpec

ExternalAccess holds configuration for the external access provided for the deployment.

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L213)

### .spec.features: DeploymentFeatures

Features allows to configure feature flags

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L170)

### .spec.id: ServerIDGroupSpec

ServerIDGroupSpec contains the specification for Image Discovery image.

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L237)

### .spec.image: string

Image specifies the docker image to use for all ArangoDB servers.
In a development environment this setting defaults to arangodb/arangodb:latest.
For production environments this is a required setting without a default value.
It is highly recommend to use explicit version (not latest) for production environments.

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L139)

### .spec.imageDiscoveryMode: string

ImageDiscoveryMode specifies the image discovery mode.

Possible Values: 
* kubelet (default) - Use sha256 of the discovered image in the pods
* direct - Use image provided in the spec.image directly in the pods

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L150)

### .spec.imagePullPolicy: core.PullPolicy

ImagePullPolicy specifies the pull policy for the docker image to use for all ArangoDB servers.

Links:
* [Documentation of core.PullPolicy](https://kubernetes.io/docs/concepts/containers/images/#image-pull-policy)

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L143)

### .spec.imagePullSecrets: []string

ImagePullSecrets specifies the list of image pull secrets for the docker image to use for all ArangoDB servers.

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L145)

### .spec.labels: map[string]string

Labels specifies the labels added to Pods in this group.

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L189)

### .spec.labelsIgnoreList: []string

LabelsIgnoreList list regexp or plain definitions which labels should be ignored

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L191)

### .spec.labelsMode: string

LabelsMode Define labels mode which should be use while overriding labels
Possible values are:
- `disabled` disable annotations/labels override. Default if there is no annotations/labels set in ArangoDeployment
- `append` add new annotations/labels without affecting old ones
- `replace` replace existing annotations/labels

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L197)

### .spec.license: LicenseSpec

License holds license settings

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L228)

### .spec.lifecycle: LifecycleSpec

Lifecycle holds lifecycle configuration settings

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L234)

### .spec.memberPropagationMode: string

MemberPropagationMode defines how changes to pod spec should be propogated.
Changes to a pod’s configuration require a restart of that pod in almost all cases.
Pods are restarted eagerly by default, which can cause more restarts than desired, especially when updating arangod as well as the operator.
The propagation of the configuration changes can be deferred to the next restart, either triggered manually by the user or by another operation like an upgrade.
This reduces the number of restarts for upgrading both the server and the operator from two to one.
- `always`: Restart the member as soon as a configuration change is discovered
- `on-restart`: Wait until the next restart to change the member configuration

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L266)

### .spec.metrics: MetricsSpec

Metrics holds metrics configuration settings

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L231)

### .spec.mode: string

Mode specifies the type of ArangoDB deployment to create.

Possible Values: 
* Cluster (default) - Full cluster. Defaults to 3 Agents, 3 DB-Servers & 3 Coordinators.
* ActiveFailover - Active-failover single pair. Defaults to 3 Agents and 2 single servers.
* Single - Single server only (note this does not provide high availability or reliability).

This field is **immutable**: Change of the ArangoDeployment Mode is not possible after creation.

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L120)

### .spec.networkAttachedVolumes: bool

NetworkAttachedVolumes
If set to `true`, a ResignLeadership operation will be triggered when a DB-Server pod is evicted (rather than a CleanOutServer operation).
Furthermore, the pod will simply be redeployed on a different node, rather than cleaned and retired and replaced by a new member.
You must only set this option to true if your persistent volumes are “movable” in the sense that they can be mounted from a different k8s node, like in the case of network attached volumes.
If your persistent volumes are tied to a specific pod, you must leave this option on false.

Default Value: false

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L177)

### .spec.rebalancer: ArangoDeploymentRebalancerSpec

Rebalancer defines the rebalancer specification

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L290)

### .spec.recovery: ArangoDeploymentRecoverySpec

Recovery specifies configuration related to cluster recovery.

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L272)

### .spec.restoreEncryptionSecret: string

RestoreEncryptionSecret specifies optional name of secret which contains encryption key used for restore

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L208)

### .spec.restoreFrom: string

RestoreFrom setting specifies a `ArangoBackup` resource name the cluster should be restored from.
After a restore or failure to do so, the status of the deployment contains information about the restore operation in the restore key.
It will contain some of the following fields:
- `requestedFrom`: name of the ArangoBackup used to restore from.
- `message`: optional message explaining why the restore failed.
- `state`: state indicating if the restore was successful or not. Possible values: Restoring, Restored, RestoreFailed
If the restoreFrom key is removed from the spec, the restore key is deleted as well.
A new restore attempt is made if and only if either in the status restore is not set or if spec.restoreFrom and status.requestedFrom are different.

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L206)

### .spec.rocksdb: RocksDBSpec

RocksDB holds rocksdb-specific configuration settings

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L216)

### .spec.single: ServerGroupSpec

Single contains specification for servers running in deployment mode `Single` or `ActiveFailover`.

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L243)

### .spec.storageEngine: string

StorageEngine specifies the type of storage engine used for all servers in the cluster.
Possible values are:
- `MMFiles` To use the MMFiles storage engine. Deprecated.
- `RocksDB` To use the RocksDB storage engine.
This setting cannot be changed after the cluster has been created.

Default Value: RocksDB

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L134)

### .spec.sync: SyncSpec

Sync holds Deployment-to-Deployment synchronization configuration settings

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L225)

### .spec.syncmasters: ServerGroupSpec

SyncMasters contains specification for Syncmaster pods running in deployment mode `Cluster`.

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L255)

### .spec.syncworkers: ServerGroupSpec

SyncWorkers contains specification for Syncworker pods running in deployment mode `Cluster`.

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L258)

### .spec.timeouts: Timeouts

Timeouts object allows to configure various time-outs

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L278)

### .spec.timezone: string

Timezone if specified, will set a timezone for deployment.
Must be in format accepted by "tzdata", e.g. `America/New_York` or `Europe/London`

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L297)

### .spec.tls: TLSSpec

TLS holds TLS configuration settings

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L222)

### .spec.topology: TopologySpec

Topology define topology adjustment details, Enterprise only

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L287)

### .spec.upgrade: DeploymentUpgradeSpec

Upgrade allows to configure upgrade-related options

[Code Reference](/pkg/apis/deployment/v1/deployment_spec.go#L167)

