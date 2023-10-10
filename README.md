# ArangoDB Kubernetes Operator

[![Docker Pulls](https://img.shields.io/docker/pulls/arangodb/kube-arangodb.svg)](https://hub.docker.com/r/arangodb/kube-arangodb/)

ArangoDB Kubernetes Operator helps to run ArangoDB deployments
on Kubernetes clusters.

To get started, follow the Installation instructions below and/or
read the [tutorial](https://www.arangodb.com/docs/stable/deployment-kubernetes-usage.html).

## State

The ArangoDB Kubernetes Operator is Production ready.

[Documentation](https://www.arangodb.com/docs/stable/deployment-kubernetes.html)

### Limits

<!-- START(limits) -->
| Limit              | Description                                                                  | Community | Enterprise |
|:-------------------|:-----------------------------------------------------------------------------|:----------|:-----------|
| Cluster size limit | Limits of the nodes (DBServers & Coordinators) supported in the Cluster mode | 64        | 1024       |

<!-- END(limits) -->

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

<!-- START(kubernetesVersionsTable) -->
| Platform            | Kubernetes Version | ArangoDB Version | State      | Remarks                                   | Provider Remarks                   |
|:--------------------|:-------------------|:-----------------|:-----------|:------------------------------------------|:-----------------------------------|
| Google GKE          | 1.21-1.26          | >= 3.6.0         | Production | Don't use micro nodes                     |                                    |
| Azure AKS           | 1.21-1.26          | >= 3.6.0         | Production |                                           |                                    |
| Amazon EKS          | 1.21-1.26          | >= 3.6.0         | Production |                                           | [Amazon EKS](./docs/providers/eks) |
| IBM Cloud           | 1.17               | >= 3.6.0         | Deprecated | Support will be dropped in Operator 1.5.0 |                                    |
| IBM Cloud           | 1.18-1.21          | >= 3.6.0         | Production |                                           |                                    |
| OpenShift           | 3.11               | >= 3.6.0         | Deprecated | Support will be dropped in Operator 1.5.0 |                                    |
| OpenShift           | 4.2-4.13           | >= 3.6.0         | Production |                                           |                                    |
| BareMetal (kubeadm) | <= 1.20            | >= 3.6.0         | Deprecated | Support will be dropped in Operator 1.5.0 |                                    |
| BareMetal (kubeadm) | 1.21-1.27          | >= 3.6.0         | Production |                                           |                                    |
| Minikube            | 1.21-1.27          | >= 3.6.0         | Devel Only |                                           |                                    |
| Other               | 1.21-1.27          | >= 3.6.0         | Devel Only |                                           |                                    |

<!-- END(kubernetesVersionsTable) -->

#### Operator Features

<!-- START(featuresCommunityTable) -->
| Feature                                                                              | Operator Version | Introduced | ArangoDB Version | ArangoDB Edition      | State        | Enabled | Flag                                                  | Remarks                                                                            |
|:-------------------------------------------------------------------------------------|:-----------------|:-----------|:-----------------|:----------------------|:-------------|:--------|:------------------------------------------------------|:-----------------------------------------------------------------------------------|
| Enforced ResignLeadership                                                            | 1.2.34           | 1.2.34     | >= 3.8.0         | Community, Enterprise | Production   | True    | --deployment.feature.enforced-resign-leadership       | Enforce ResignLeadership and ensure that Leaders are moved from restarted DBServer |
| Copy resources spec to init containers                                               | 1.2.34           | 1.2.34     | >= 3.8.0         | Community, Enterprise | Production   | True    | --deployment.feature.init-containers-copy-resources   | Copy resources spec to built-in init containers if they are not specified          |
| [Rebalancer V2](docs/design/features/rebalancer_v2.md)                               | 1.2.31           | 1.2.31     | >= 3.10.0        | Community, Enterprise | Alpha        | False   | --deployment.feature.rebalancer-v2                    | N/A                                                                                |
| [Secured containers](docs/design/features/secured_containers.md)                     | 1.2.31           | 1.2.31     | >= 3.8.0         | Community, Enterprise | Alpha        | False   | --deployment.feature.secured-containers               | If set to True Operator will run containers in secure mode                         |
| Version Check V2                                                                     | 1.2.31           | 1.2.31     | >= 3.8.0         | Community, Enterprise | Alpha        | False   | --deployment.feature.upgrade-version-check-V2         | N/A                                                                                |
| [Operator Ephemeral Volumes](docs/design/features/ephemeral_volumes.md)              | 1.2.31           | 1.2.2      | >= 3.8.0         | Community, Enterprise | Beta         | False   | --deployment.feature.ephemeral-volumes                | N/A                                                                                |
| [Force Rebuild Out Synced Shards](docs/design/features/rebuild_out_synced_shards.md) | 1.2.27           | 1.2.27     | >= 3.8.0         | Community, Enterprise | Production   | False   | --deployment.feature.force-rebuild-out-synced-shards  | It should be used only if user is aware of the risks.                              |
| [Spec Default Restore](docs/design/features/deployment_spec_defaults.md)             | 1.2.25           | 1.2.21     | >= 3.8.0         | Community, Enterprise | Beta         | True    | --deployment.feature.deployment-spec-defaults-restore | If set to False Operator will not change ArangoDeployment Spec                     |
| Version Check                                                                        | 1.2.23           | 1.1.4      | >= 3.8.0         | Community, Enterprise | Production   | True    | --deployment.feature.upgrade-version-check            | N/A                                                                                |
| [Failover Leader service](docs/design/features/failover_leader_service.md)           | 1.2.13           | 1.2.13     | >= 3.8.0         | Community, Enterprise | Production   | False   | --deployment.feature.failover-leadership              | N/A                                                                                |
| Graceful Restart                                                                     | 1.2.5            | 1.0.7      | >= 3.8.0         | Community, Enterprise | Production   | True    | ---deployment.feature.graceful-shutdown               | N/A                                                                                |
| Optional Graceful Restart                                                            | 1.2.0            | 1.2.5      | >= 3.8.0         | Community, Enterprise | Production   | False   | --deployment.feature.optional-graceful-shutdown       | N/A                                                                                |
| Operator Internal Metrics Exporter                                                   | 1.2.0            | 1.2.0      | >= 3.8.0         | Community, Enterprise | Production   | True    | --deployment.feature.metrics-exporter                 | N/A                                                                                |
| Operator Maintenance Management Support                                              | 1.2.0            | 1.0.7      | >= 3.8.0         | Community, Enterprise | Production   | True    | --deployment.feature.maintenance                      | N/A                                                                                |
| Encryption Key Rotation Support                                                      | 1.2.0            | 1.0.3      | >= 3.8.0         | Enterprise            | NotSupported | False   | --deployment.feature.encryption-rotation              | N/A                                                                                |
| TLS Runtime Rotation Support                                                         | 1.1.0            | 1.0.4      | >= 3.8.0         | Enterprise            | Production   | True    | --deployment.feature.tls-rotation                     | N/A                                                                                |
| JWT Rotation Support                                                                 | 1.1.0            | 1.0.3      | >= 3.8.0         | Enterprise            | Production   | True    | --deployment.feature.jwt-rotation                     | N/A                                                                                |
| Operator Single Mode                                                                 | 1.0.4            | 1.0.4      | >= 3.8.0         | Community, Enterprise | Production   | False   | --mode.single                                         | Only 1 instance of Operator allowed in namespace when feature is enabled           |
| TLS SNI Support                                                                      | 1.0.3            | 1.0.3      | >= 3.8.0         | Enterprise            | Production   | True    | --deployment.feature.tls-sni                          | N/A                                                                                |
| Disabling of liveness probes                                                         | 0.3.11           | 0.3.10     | >= 3.8.0         | Community, Enterprise | Production   | True    | N/A                                                   | N/A                                                                                |
| Pod Disruption Budgets                                                               | 0.3.11           | 0.3.10     | >= 3.8.0         | Community, Enterprise | Production   | True    | N/A                                                   | N/A                                                                                |
| Prometheus Metrics Exporter                                                          | 0.3.11           | 0.3.10     | >= 3.8.0         | Community, Enterprise | Production   | True    | N/A                                                   | Prometheus required                                                                |
| Sidecar Containers                                                                   | 0.3.11           | 0.3.10     | >= 3.8.0         | Community, Enterprise | Production   | True    | N/A                                                   | N/A                                                                                |
| Volume Claim Templates                                                               | 0.3.11           | 0.3.10     | >= 3.8.0         | Community, Enterprise | Production   | True    | N/A                                                   | N/A                                                                                |
| Volume Resizing                                                                      | 0.3.11           | 0.3.10     | >= 3.8.0         | Community, Enterprise | Production   | True    | N/A                                                   | N/A                                                                                |

<!-- END(featuresCommunityTable) -->

#### Operator Enterprise Only Features

To upgrade to the Enterprise Edition, you need to get in touch with the ArangoDB team. [Contact us](https://www.arangodb.com/contact/) for more details.

<!-- START(featuresEnterpriseTable) -->
| Feature                                                | Operator Version | Introduced | ArangoDB Version | ArangoDB Edition | State      | Enabled | Flag | Remarks                                                                     |
|:-------------------------------------------------------|:-----------------|:-----------|:-----------------|:-----------------|:-----------|:--------|:-----|:----------------------------------------------------------------------------|
| AgencyCache                                            | 1.2.30           | 1.2.30     | >= 3.8.0         | Enterprise       | Production | True    | N/A  | Enable Agency Cache mechanism in the Operator (Increase limit of the nodes) |
| Member Maintenance Support                             | 1.2.25           | 1.2.16     | >= 3.8.0         | Enterprise       | Production | True    | N/A  | Enable Member Maintenance during planned restarts                           |
| [Rebalancer](docs/design/features/rebalancer.md)       | 1.2.15           | 1.2.5      | >= 3.8.0         | Enterprise       | Production | True    | N/A  | N/A                                                                         |
| [TopologyAwareness](docs/design/topology_awareness.md) | 1.2.4            | 1.2.4      | >= 3.8.0         | Enterprise       | Production | True    | N/A  | N/A                                                                         |

<!-- END(featuresEnterpriseTable) -->

## Installation and Usage

Docker images:
- Community Edition: `arangodb/kube-arangodb:1.2.34`
- Enterprise Edition: `arangodb/kube-arangodb-enterprise:1.2.34`

### Installation of latest release using Kubectl

This procedure can also be used for upgrades and will not harm any
running ArangoDB deployments.

##### Community Edition
```bash
kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/1.2.34/manifests/arango-crd.yaml
kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/1.2.34/manifests/arango-deployment.yaml
# To use `ArangoLocalStorage`, also run
kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/1.2.34/manifests/arango-storage.yaml
# To use `ArangoDeploymentReplication`, also run
kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/1.2.34/manifests/arango-deployment-replication.yaml
```

##### Enterprise Edition
```bash
kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/1.2.34/manifests/enterprise-crd.yaml
kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/1.2.34/manifests/enterprise-deployment.yaml
# To use `ArangoLocalStorage`, also run
kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/1.2.34/manifests/enterprise-storage.yaml
# To use `ArangoDeploymentReplication`, also run
kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/1.2.34/manifests/enterprise-deployment-replication.yaml
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
  - https://github.com/arangodb/kube-arangodb/manifests/kustomize/crd?ref=1.2.33
  - https://github.com/arangodb/kube-arangodb/manifests/kustomize/deployment?ref=1.2.33
```

##### Enterprise Edition example
```
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: my-custom-namespace
resources:
  - https://github.com/arangodb/kube-arangodb/manifests/kustomize-enterprise/crd?ref=1.2.33
  - https://github.com/arangodb/kube-arangodb/manifests/kustomize-enterprise/deployment?ref=1.2.33
```

### Installation of latest release using Helm

Only use this procedure for clean installation of the operator. For upgrades see next section

##### Community Edition
```bash
# The following will install the operator and basic CRDs resources.
helm install --generate-name https://github.com/arangodb/kube-arangodb/releases/download/1.2.34/kube-arangodb-1.2.34.tgz
# To use `ArangoLocalStorage`, set field `operator.features.storage` to true
helm install --generate-name https://github.com/arangodb/kube-arangodb/releases/download/1.2.34/kube-arangodb-1.2.34.tgz --set "operator.features.storage=true"
```

##### Enterprise Edition
```bash
# The following will install the operator and basic CRDs resources.
helm install --generate-name https://github.com/arangodb/kube-arangodb/releases/download/1.2.34/kube-arangodb-1.2.34.tgz --set "operator.image=arangodb/kube-arangodb-enterprise:1.2.34"
# To use `ArangoLocalStorage`, set field `operator.features.storage` to true
helm install --generate-name https://github.com/arangodb/kube-arangodb/releases/download/1.2.34/kube-arangodb-1.2.34.tgz --set "operator.image=arangodb/kube-arangodb-enterprise:1.2.34" --set "operator.features.storage=true"
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
helm install --generate-name https://github.com/arangodb/kube-arangodb/releases/download/1.2.34/kube-arangodb-1.2.34.tgz
# To use `ArangoLocalStorage`, set field `operator.features.storage` to true
helm install --generate-name https://github.com/arangodb/kube-arangodb/releases/download/1.2.34/kube-arangodb-1.2.34.tgz --set "operator.features.storage=true"
```

##### Enterprise Edition
```bash
# The following will install the operator and basic CRDs resources.
helm install --generate-name https://github.com/arangodb/kube-arangodb/releases/download/1.2.34/kube-arangodb-1.2.34.tgz --set "operator.image=arangodb/kube-arangodb-enterprise:1.2.34"
# To use `ArangoLocalStorage`, set field `operator.features.storage` to true
helm install --generate-name https://github.com/arangodb/kube-arangodb/releases/download/1.2.34/kube-arangodb-1.2.34.tgz --set "operator.image=arangodb/kube-arangodb-enterprise:1.2.34" --set "operator.features.storage=true"
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
