# ArangoDB Kubernetes Operator

[![Docker Pulls](https://img.shields.io/docker/pulls/arangodb/kube-arangodb.svg)](https://hub.docker.com/r/arangodb/kube-arangodb/) [![CircleCI](https://dl.circleci.com/status-badge/img/gh/arangodb/kube-arangodb/tree/master.svg?style=svg)](https://dl.circleci.com/status-badge/redirect/gh/arangodb/kube-arangodb/tree/master)

The ArangoDB Kubernetes Operator (`kube-arangodb`) is a set of operators
that you deploy in your Kubernetes cluster to:
- Manage deployments of the [ArangoDB database](https://arangodb.com/)
- Manage backups
- Provide `PersistentVolumes` on local storage of your nodes for optimal storage performance.
- Configure ArangoDB Datacenter-to-Datacenter Replication

Each of these uses involves a different custom resource:
- Use an [ArangoDeployment resource](docs/deployment-resource-reference.md) to create an ArangoDB database deployment.
- Use an [ArangoMember resource](docs/api/ArangoMember.V1.md) to observe and adjust individual deployment members.
- Use an [ArangoBackup](docs/backup-resource.md) and [ArangoBackupPolicy](docs/backuppolicy-resource.md) resources to create ArangoDB backups.
- Use an [ArangoLocalStorage resource](docs/storage-resource.md) to provide local `PersistentVolumes` for optimal I/O performance.
- Use an [ArangoDeploymentReplication resource](docs/deployment-replication-resource-reference.md) to configure ArangoDB Datacenter-to-Datacenter Replication.

Continue with [Using the ArangoDB Kubernetes Operator](docs/using-the-operator.md)
to learn how to install the ArangoDB Kubernetes operator and create your first deployment.

## State

The ArangoDB Kubernetes Operator is Production ready.

[Documentation](https://arangodb.github.io/kube-arangodb/)

### Limits

[START_INJECT]: # (limits)

| Limit              | Description                                                                  | Community | Enterprise |
|:-------------------|:-----------------------------------------------------------------------------|:----------|:-----------|
| Cluster size limit | Limits of the nodes (DBServers & Coordinators) supported in the Cluster mode | 64        | 1024       |

[END_INJECT]: # (limits)

### Production readiness state

Beginning with Version 0.3.11 we maintain a production readiness
state for individual new features, since we expect that new
features will first be released with an "alpha" or "beta" readiness
state and over time move to full "production readiness".

Operator will support Kubernetes versions supported on providers and maintained by Kubernetes.
Once version is not supported anymore it will go into "Deprecating" state and will be marked as deprecated on Minor release.

Kubernetes versions starting from 1.18 are supported and tested, charts and manifests can use API Versions which are not present in older versions.

The following table has the general readiness state, the table below
covers individual newer features separately.

[START_INJECT]: # (kubernetesVersionsTable)

| Platform            | Kubernetes Version | ArangoDB Version | State      | Remarks                                   | Provider Remarks                   |
|:--------------------|:-------------------|:-----------------|:-----------|:------------------------------------------|:-----------------------------------|
| Google GKE          | 1.21-1.28          | >= 3.8.0         | Production | Don't use micro nodes                     |                                    |
| Azure AKS           | 1.21-1.28          | >= 3.8.0         | Production |                                           |                                    |
| Amazon EKS          | 1.21-1.28          | >= 3.8.0         | Production |                                           | [Amazon EKS](./docs/providers/eks) |
| IBM Cloud           | <= 1.20            | >= 3.8.0         | Deprecated | Support will be dropped in Operator 1.5.0 |                                    |
| IBM Cloud           | 1.21-1.28          | >= 3.8.0         | Production |                                           |                                    |
| OpenShift           | 3.11               | >= 3.8.0         | Deprecated | Support will be dropped in Operator 1.5.0 |                                    |
| OpenShift           | 4.2-4.14           | >= 3.8.0         | Production |                                           |                                    |
| BareMetal (kubeadm) | <= 1.20            | >= 3.8.0         | Deprecated | Support will be dropped in Operator 1.5.0 |                                    |
| BareMetal (kubeadm) | 1.21-1.28          | >= 3.8.0         | Production |                                           |                                    |
| Minikube            | 1.21-1.28          | >= 3.8.0         | Devel Only |                                           |                                    |
| Other               | 1.21-1.28          | >= 3.8.0         | Devel Only |                                           |                                    |

[END_INJECT]: # (kubernetesVersionsTable)

#### Operator Features

[START_INJECT]: # (featuresCommunityTable)

| Feature                                                                       | Operator Version | Introduced | ArangoDB Version | ArangoDB Edition      | State        | Enabled | Flag                                                   | Remarks                                                                               |
|:------------------------------------------------------------------------------|:-----------------|:-----------|:-----------------|:----------------------|:-------------|:--------|:-------------------------------------------------------|:--------------------------------------------------------------------------------------|
| Upscale resources spec in init containers                                     | 1.2.36           | 1.2.36     | >= 3.8.0         | Community, Enterprise | Production   | True    | --deployment.feature.init-containers-upscale-resources | Upscale resources spec to built-in init containers if they are not specified or lower |
| Create backups asynchronously                                                 | 1.2.35           | 1.2.35     | >= 3.8.0         | Community, Enterprise | Production   | False   | --deployment.feature.async-backup-creation             | Create backups asynchronously to avoid blocking the operator and reaching the timeout |
| Enforced ResignLeadership                                                     | 1.2.34           | 1.2.34     | >= 3.8.0         | Community, Enterprise | Production   | True    | --deployment.feature.enforced-resign-leadership        | Enforce ResignLeadership and ensure that Leaders are moved from restarted DBServer    |
| Copy resources spec to init containers                                        | 1.2.33           | 1.2.33     | >= 3.8.0         | Community, Enterprise | Production   | True    | --deployment.feature.init-containers-copy-resources    | Copy resources spec to built-in init containers if they are not specified             |
| [Rebalancer V2](docs/features/rebalancer_v2.md)                               | 1.2.31           | 1.2.31     | >= 3.10.0        | Community, Enterprise | Alpha        | False   | --deployment.feature.rebalancer-v2                     | N/A                                                                                   |
| [Secured containers](docs/features/secured_containers.md)                     | 1.2.31           | 1.2.31     | >= 3.8.0         | Community, Enterprise | Alpha        | False   | --deployment.feature.secured-containers                | If set to True Operator will run containers in secure mode                            |
| Version Check V2                                                              | 1.2.31           | 1.2.31     | >= 3.8.0         | Community, Enterprise | Alpha        | False   | --deployment.feature.upgrade-version-check-V2          | N/A                                                                                   |
| [Operator Ephemeral Volumes](docs/features/ephemeral_volumes.md)              | 1.2.31           | 1.2.2      | >= 3.8.0         | Community, Enterprise | Beta         | False   | --deployment.feature.ephemeral-volumes                 | N/A                                                                                   |
| [Force Rebuild Out Synced Shards](docs/features/rebuild_out_synced_shards.md) | 1.2.27           | 1.2.27     | >= 3.8.0         | Community, Enterprise | Production   | False   | --deployment.feature.force-rebuild-out-synced-shards   | It should be used only if user is aware of the risks.                                 |
| [Spec Default Restore](docs/features/deployment_spec_defaults.md)             | 1.2.25           | 1.2.21     | >= 3.8.0         | Community, Enterprise | Beta         | True    | --deployment.feature.deployment-spec-defaults-restore  | If set to False Operator will not change ArangoDeployment Spec                        |
| Version Check                                                                 | 1.2.23           | 1.1.4      | >= 3.8.0         | Community, Enterprise | Production   | True    | --deployment.feature.upgrade-version-check             | N/A                                                                                   |
| [Failover Leader service](docs/features/failover_leader_service.md)           | 1.2.13           | 1.2.13     | < 3.12.0         | Community, Enterprise | Production   | False   | --deployment.feature.failover-leadership               | N/A                                                                                   |
| Graceful Restart                                                              | 1.2.5            | 1.0.7      | >= 3.8.0         | Community, Enterprise | Production   | True    | ---deployment.feature.graceful-shutdown                | N/A                                                                                   |
| Optional Graceful Restart                                                     | 1.2.0            | 1.2.5      | >= 3.8.0         | Community, Enterprise | Production   | False   | --deployment.feature.optional-graceful-shutdown        | N/A                                                                                   |
| Operator Internal Metrics Exporter                                            | 1.2.0            | 1.2.0      | >= 3.8.0         | Community, Enterprise | Production   | True    | --deployment.feature.metrics-exporter                  | N/A                                                                                   |
| Operator Maintenance Management Support                                       | 1.2.0            | 1.0.7      | >= 3.8.0         | Community, Enterprise | Production   | True    | --deployment.feature.maintenance                       | N/A                                                                                   |
| Encryption Key Rotation Support                                               | 1.2.0            | 1.0.3      | >= 3.8.0         | Enterprise            | NotSupported | False   | --deployment.feature.encryption-rotation               | N/A                                                                                   |
| TLS Runtime Rotation Support                                                  | 1.1.0            | 1.0.4      | >= 3.8.0         | Enterprise            | Production   | True    | --deployment.feature.tls-rotation                      | N/A                                                                                   |
| JWT Rotation Support                                                          | 1.1.0            | 1.0.3      | >= 3.8.0         | Enterprise            | Production   | True    | --deployment.feature.jwt-rotation                      | N/A                                                                                   |
| Operator Single Mode                                                          | 1.0.4            | 1.0.4      | >= 3.8.0         | Community, Enterprise | Production   | False   | --mode.single                                          | Only 1 instance of Operator allowed in namespace when feature is enabled              |
| TLS SNI Support                                                               | 1.0.3            | 1.0.3      | >= 3.8.0         | Enterprise            | Production   | True    | --deployment.feature.tls-sni                           | N/A                                                                                   |
| ActiveFailover Support                                                        | 1.0.0            | 1.0.0      | < 3.12.0         | Community, Enterprise | Production   | True    | --deployment.feature.active-failover                   | N/A                                                                                   |
| Disabling of liveness probes                                                  | 0.3.11           | 0.3.10     | >= 3.8.0         | Community, Enterprise | Production   | True    | N/A                                                    | N/A                                                                                   |
| Pod Disruption Budgets                                                        | 0.3.11           | 0.3.10     | >= 3.8.0         | Community, Enterprise | Production   | True    | N/A                                                    | N/A                                                                                   |
| Prometheus Metrics Exporter                                                   | 0.3.11           | 0.3.10     | >= 3.8.0         | Community, Enterprise | Production   | True    | N/A                                                    | Prometheus required                                                                   |
| Sidecar Containers                                                            | 0.3.11           | 0.3.10     | >= 3.8.0         | Community, Enterprise | Production   | True    | N/A                                                    | N/A                                                                                   |
| Volume Claim Templates                                                        | 0.3.11           | 0.3.10     | >= 3.8.0         | Community, Enterprise | Production   | True    | N/A                                                    | N/A                                                                                   |
| Volume Resizing                                                               | 0.3.11           | 0.3.10     | >= 3.8.0         | Community, Enterprise | Production   | True    | N/A                                                    | N/A                                                                                   |

[END_INJECT]: # (featuresCommunityTable)

#### Operator Enterprise Only Features

To upgrade to the Enterprise Edition, you need to get in touch with the ArangoDB team. [Contact us](https://www.arangodb.com/contact/) for more details.

[START_INJECT]: # (featuresEnterpriseTable)

| Feature                                                | Operator Version | Introduced | ArangoDB Version | ArangoDB Edition | State      | Enabled | Flag | Remarks                                                                     |
|:-------------------------------------------------------|:-----------------|:-----------|:-----------------|:-----------------|:-----------|:--------|:-----|:----------------------------------------------------------------------------|
| ArangoML integration                                   | 1.2.36           | 1.2.36     | >= 3.8.0         | Enterprise       | Alpha      | True    | N/A  | Support for ArangoML CRDs                                                   |
| AgencyCache                                            | 1.2.30           | 1.2.30     | >= 3.8.0         | Enterprise       | Production | True    | N/A  | Enable Agency Cache mechanism in the Operator (Increase limit of the nodes) |
| Member Maintenance Support                             | 1.2.25           | 1.2.16     | >= 3.8.0         | Enterprise       | Production | True    | N/A  | Enable Member Maintenance during planned restarts                           |
| [Rebalancer](docs/features/rebalancer.md)              | 1.2.15           | 1.2.5      | >= 3.8.0         | Enterprise       | Production | True    | N/A  | N/A                                                                         |
| [TopologyAwareness](docs/design/topology_awareness.md) | 1.2.4            | 1.2.4      | >= 3.8.0         | Enterprise       | Production | True    | N/A  | N/A                                                                         |

[END_INJECT]: # (featuresEnterpriseTable)

## Flags

```
Flags:
      --action.PVCResize.concurrency int                       Define limit of concurrent PVC Resizes on the cluster (default 32)
      --agency.refresh-delay duration                          The Agency refresh delay (0 = no delay) (default 500ms)
      --agency.refresh-interval duration                       The Agency refresh interval (0 = do not refresh)
      --agency.retries int                                     The Agency retries (0 = no retries) (default 1)
      --api.enabled                                            Enable operator HTTP and gRPC API (default true)
      --api.grpc-port int                                      gRPC API port to listen on (default 8728)
      --api.http-port int                                      HTTP API port to listen on (default 8628)
      --api.jwt-key-secret-name string                         Name of secret containing key used to sign JWT. If there is no such secret present, value will be saved here (default "arangodb-operator-api-jwt-key")
      --api.jwt-secret-name string                             Name of secret which will contain JWT to authenticate API requests. (default "arangodb-operator-api-jwt")
      --api.tls-secret-name string                             Name of secret containing tls.crt & tls.key for HTTPS API (if empty, self-signed certificate is used)
      --backup-concurrent-uploads int                          Number of concurrent uploads per deployment (default 4)
      --chaos.allowed                                          Set to allow chaos in deployments. Only activated when allowed and enabled in deployment
      --crd.install                                            Install missing CRD if access is possible (default true)
      --crd.validation-schema stringArray                      Overrides default set of CRDs which should have validation schema enabled <crd-name>=<true/false>.
      --deployment.feature.active-failover                     Support for ActiveFailover mode - Required ArangoDB >= 3.8.0, < 3.12 (default true)
      --deployment.feature.agency-poll                         Enable Agency Poll for Enterprise deployments - Required ArangoDB >= 3.8.0 (default true)
      --deployment.feature.all                                 Enable ALL Features
      --deployment.feature.async-backup-creation               Create backups asynchronously to avoid blocking the operator and reaching the timeout - Required ArangoDB >= 3.8.0
      --deployment.feature.deployment-spec-defaults-restore    Restore defaults from last accepted state of deployment - Required ArangoDB >= 3.8.0 (default true)
      --deployment.feature.enforced-resign-leadership          Enforce ResignLeadership and ensure that Leaders are moved from restarted DBServer - Required ArangoDB >= 3.8.0 (default true)
      --deployment.feature.ephemeral-volumes                   Enables ephemeral volumes for apps and tmp directory - Required ArangoDB >= 3.8.0
      --deployment.feature.failover-leadership                 Support for leadership in fail-over mode - Required ArangoDB >= 3.8.0, < 3.12
      --deployment.feature.init-containers-copy-resources      Copy resources spec to built-in init containers if they are not specified - Required ArangoDB >= 3.8.0 (default true)
      --deployment.feature.init-containers-upscale-resources   Copy resources spec to built-in init containers if they are not specified or lower - Required ArangoDB >= 3.8.0 (default true)
      --deployment.feature.local-storage.pass-reclaim-policy   [LocalStorage] Pass ReclaimPolicy from StorageClass instead of using hardcoded Retain - Required ArangoDB >= 3.8.0
      --deployment.feature.local-volume-replacement-check      Replace volume for local-storage if volume is unschedulable (ex. node is gone) - Required ArangoDB >= 3.8.0
      --deployment.feature.random-pod-names                    Enables generating random pod names - Required ArangoDB >= 3.8.0
      --deployment.feature.rebalancer-v2                       Rebalancer V2 feature - Required ArangoDB >= 3.10.0
      --deployment.feature.restart-policy-always               Allow to restart containers with always restart policy - Required ArangoDB >= 3.8.0
      --deployment.feature.secured-containers                  Create server's containers with non root privileges. It enables 'ephemeral-volumes' feature implicitly - Required ArangoDB >= 3.8.0
      --deployment.feature.sensitive-information-protection    Hide sensitive information from metrics and logs - Required ArangoDB >= 3.8.0
      --deployment.feature.short-pod-names                     Enable Short Pod Names - Required ArangoDB >= 3.8.0
      --deployment.feature.timezone-management                 Enable timezone management for pods - Required ArangoDB >= 3.8.0
      --deployment.feature.tls-sni                             TLS SNI Support - Required ArangoDB EE >= 3.8.0 (default true)
      --deployment.feature.upgrade-version-check               Enable initContainer with pre version check - Required ArangoDB >= 3.8.0 (default true)
      --deployment.feature.upgrade-version-check-v2            Enable initContainer with pre version check based by Operator - Required ArangoDB >= 3.8.0
      --features-config-map-name string                        Name of the Feature Map ConfigMap (default "arangodb-operator-feature-config-map")
  -h, --help                                                   help for arangodb_operator
      --image.discovery.status                                 Discover Operator Image from Pod Status by default. When disabled Pod Spec is used. (default true)
      --image.discovery.timeout duration                       Timeout for image discovery process (default 1m0s)
      --internal.scaling-integration                           Enable Scaling Integration
      --kubernetes.burst int                                   Burst for the k8s API (default 30)
      --kubernetes.max-batch-size int                          Size of batch during objects read (default 256)
      --kubernetes.qps float32                                 Number of queries per second for k8s API (default 15)
      --log.format string                                      Set log format. Allowed values: 'pretty', 'JSON'. If empty, default format is used (default "pretty")
      --log.level stringArray                                  Set log levels in format <level> or <logger>=<level>. Possible loggers: action, agency, api-server, assertion, backup-operator, chaos-monkey, crd, deployment, deployment-ci, deployment-reconcile, deployment-replication, deployment-resilience, deployment-resources, deployment-storage, deployment-storage-pc, deployment-storage-service, http, inspector, integrations, k8s-client, ml-batchjob-operator, ml-cronjob-operator, ml-extension-operator, ml-extension-shutdown, ml-storage-operator, monitor, operator, operator-arangojob-handler, operator-v2, operator-v2-event, operator-v2-worker, panics, pod_compare, root, root-event-recorder, scheduler, server, server-authentication (default [info])
      --log.sampling                                           If true, operator will try to minimize duplication of logging events (default true)
      --memory-limit uint                                      Define memory limit for hard shutdown and the dump of goroutines. Used for testing
      --metrics.excluded-prefixes stringArray                  List of the excluded metrics prefixes
      --mode.single                                            Enable single mode in Operator. WARNING: There should be only one replica of Operator, otherwise Operator can take unexpected actions
      --operator.apps                                          Enable to run the ArangoApps operator
      --operator.backup                                        Enable to run the ArangoBackup operator
      --operator.deployment                                    Enable to run the ArangoDeployment operator
      --operator.deployment-replication                        Enable to run the ArangoDeploymentReplication operator
      --operator.ml                                            Enable to run the ArangoML operator
      --operator.reconciliation.retry.count int                Count of retries during Object Update operations in the Reconciliation loop (default 25)
      --operator.reconciliation.retry.delay duration           Delay between Object Update operations in the Reconciliation loop (default 1s)
      --operator.storage                                       Enable to run the ArangoLocalStorage operator
      --operator.version                                       Enable only version endpoint in Operator
      --reconciliation.delay duration                          Delay between reconciliation loops (<= 0 -> Disabled)
      --scope string                                           Define scope on which Operator works. Legacy - pre 1.1.0 scope with limited cluster access (default "legacy")
      --server.admin-secret-name string                        Name of secret containing username + password for login to the dashboard (default "arangodb-operator-dashboard")
      --server.allow-anonymous-access                          Allow anonymous access to the dashboard
      --server.host string                                     Host to listen on (default "0.0.0.0")
      --server.port int                                        Port to listen on (default 8528)
      --server.tls-secret-name string                          Name of secret containing tls.crt & tls.key for HTTPS server (if empty, self-signed certificate is used)
      --shutdown.delay duration                                The delay before running shutdown handlers (default 2s)
      --shutdown.timeout duration                              Timeout for shutdown handlers (default 30s)
      --timeout.agency duration                                The Agency read timeout (default 10s)
      --timeout.arangod duration                               The request timeout to the ArangoDB (default 5s)
      --timeout.arangod-check duration                         The version check request timeout to the ArangoDB (default 2s)
      --timeout.backup-arangod duration                        The request timeout to the ArangoDB during backup calls (default 30s)
      --timeout.backup-upload duration                         The request timeout to the ArangoDB during uploading files (default 5m0s)
      --timeout.force-delete-pod-grace-period duration         Default period when ArangoDB Pod should be forcefully removed after all containers were stopped - set to 0 to disable forceful removals (default 15m0s)
      --timeout.k8s duration                                   The request timeout to the kubernetes (default 2s)
      --timeout.reconciliation duration                        The reconciliation timeout to the ArangoDB CR (default 1m0s)
      --timeout.shard-rebuild duration                         Timeout after which particular out-synced shard is considered as failed and rebuild is triggered (default 1h0m0s)
      --timeout.shard-rebuild-retry duration                   Timeout after which rebuild shards retry flow is triggered (default 4h0m0s)
```

## Installation and Usage

Docker images:
- Community Edition: `arangodb/kube-arangodb:1.2.39`
- Enterprise Edition: `arangodb/kube-arangodb-enterprise:1.2.39`

### Installation of latest release using Kubectl

This procedure can also be used for upgrades and will not harm any
running ArangoDB deployments.

##### Community Edition
```bash
kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/1.2.39/manifests/arango-crd.yaml
kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/1.2.39/manifests/arango-deployment.yaml
# To use `ArangoLocalStorage`, also run
kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/1.2.39/manifests/arango-storage.yaml
# To use `ArangoDeploymentReplication`, also run
kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/1.2.39/manifests/arango-deployment-replication.yaml
```

##### Enterprise Edition
```bash
kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/1.2.39/manifests/enterprise-crd.yaml
kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/1.2.39/manifests/enterprise-deployment.yaml
# To use `ArangoLocalStorage`, also run
kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/1.2.39/manifests/enterprise-storage.yaml
# To use `ArangoDeploymentReplication`, also run
kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/1.2.39/manifests/enterprise-deployment-replication.yaml
```

### Installation of latest release using kustomize

Installation using [kustomize](https://kubernetes.io/docs/tasks/manage-kubernetes-objects/kustomization/) looks like installation from yaml files,
but user is allowed to modify namespace or resource names without yaml modifications.

It is recommended to use kustomization instead of handcrafting namespace in yaml files - kustomization will replace not only resource namespaces,
but also namespace references in resources like ClusterRoleBinding.

See `manifests/kustomize` directory for available combinations of installed features.

##### Community Edition example
```
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: my-custom-namespace
resources:
  - https://github.com/arangodb/kube-arangodb/manifests/kustomize/crd?ref=1.2.36
  - https://github.com/arangodb/kube-arangodb/manifests/kustomize/deployment?ref=1.2.36
```

##### Enterprise Edition example
```
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: my-custom-namespace
resources:
  - https://github.com/arangodb/kube-arangodb/manifests/kustomize-enterprise/crd?ref=1.2.36
  - https://github.com/arangodb/kube-arangodb/manifests/kustomize-enterprise/deployment?ref=1.2.36
```

### Installation of latest release using Helm

Only use this procedure for clean installation of the operator. For upgrades see next section

##### Community Edition
```bash
# The following will install the operator and basic CRDs resources.
helm install --generate-name https://github.com/arangodb/kube-arangodb/releases/download/1.2.39/kube-arangodb-1.2.39.tgz
# To use `ArangoLocalStorage`, set field `operator.features.storage` to true
helm install --generate-name https://github.com/arangodb/kube-arangodb/releases/download/1.2.39/kube-arangodb-1.2.39.tgz --set "operator.features.storage=true"
```

##### Enterprise Edition
```bash
# The following will install the operator and basic CRDs resources.
helm install --generate-name https://github.com/arangodb/kube-arangodb/releases/download/1.2.39/kube-arangodb-1.2.39.tgz --set "operator.image=arangodb/kube-arangodb-enterprise:1.2.39"
# To use `ArangoLocalStorage`, set field `operator.features.storage` to true
helm install --generate-name https://github.com/arangodb/kube-arangodb/releases/download/1.2.39/kube-arangodb-1.2.39.tgz --set "operator.image=arangodb/kube-arangodb-enterprise:1.2.39" --set "operator.features.storage=true"
```

### Upgrading the operator using Helm

To upgrade the operator to the latest version with Helm, you have to
delete the previous operator deployment and then install the latest. **HOWEVER**:
You *must not delete* the custom resource definitions (CRDs),
or your ArangoDB deployments will be deleted!

Therefore, you have to use `helm list` to find the deployments for the
operator (`kube-arangodb`) and use `helm delete` to delete them using the
automatically generated deployment names. Here is an example of a `helm list` output:

```
NAME                      	NAMESPACE	REVISION	UPDATED                                 	STATUS  	CHART               	APP VERSION
kube-arangodb-1-1696919877	default  	1       	2023-10-10 08:37:57.884783199 +0200 CEST	deployed	kube-arangodb-1.2.31	
```

So here, you would have to do
```bash
helm delete kube-arangodb-1-1696919877
```

Then you can install the new version with `helm install` as normal:

##### Community Edition
```bash
# The following will install the operator and basic CRDs resources.
helm install --generate-name https://github.com/arangodb/kube-arangodb/releases/download/1.2.39/kube-arangodb-1.2.39.tgz
# To use `ArangoLocalStorage`, set field `operator.features.storage` to true
helm install --generate-name https://github.com/arangodb/kube-arangodb/releases/download/1.2.39/kube-arangodb-1.2.39.tgz --set "operator.features.storage=true"
```

##### Enterprise Edition
```bash
# The following will install the operator and basic CRDs resources.
helm install --generate-name https://github.com/arangodb/kube-arangodb/releases/download/1.2.39/kube-arangodb-1.2.39.tgz --set "operator.image=arangodb/kube-arangodb-enterprise:1.2.39"
# To use `ArangoLocalStorage`, set field `operator.features.storage` to true
helm install --generate-name https://github.com/arangodb/kube-arangodb/releases/download/1.2.39/kube-arangodb-1.2.39.tgz --set "operator.image=arangodb/kube-arangodb-enterprise:1.2.39" --set "operator.features.storage=true"
```

## Building

```bash
DOCKERNAMESPACE=<your dockerhub account> make
kubectl apply -f manifests/arango-deployment-dev.yaml
# To use `ArangoLocalStorage`, also run
kubectl apply -f manifests/arango-storage-dev.yaml
# To use `ArangoDeploymentReplication`, also run
kubectl apply -f manifests/arango-deployment-replication-dev.yaml
```
