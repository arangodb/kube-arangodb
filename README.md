# ArangoDB Kubernetes Operator

[![Docker Pulls](https://img.shields.io/docker/pulls/arangodb/kube-arangodb.svg)](https://hub.docker.com/r/arangodb/kube-arangodb/)

ArangoDB Kubernetes Operator helps to run ArangoDB deployments
on Kubernetes clusters.

To get started, follow the Installation instructions below and/or
read the [tutorial](https://www.arangodb.com/docs/stable/deployment-kubernetes-usage.html).

## State

The ArangoDB Kubernetes Operator is Production ready.

[Documentation](https://www.arangodb.com/docs/stable/deployment-kubernetes.html)

### Production readiness state

Beginning with Version 0.3.11 we maintain a production readiness
state for individual new features, since we expect that new
features will first be released with an "alpha" or "beta" readiness
state and over time move to full "production readiness".

Operator will supports versions supported on providers and maintained by Kubernetes.
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

#### Feature-wise production readiness table:

<!-- START(featuresTable) -->
| Feature                                                                               | Operator Version | Operator Edition      | Introduced | ArangoDB Version | ArangoDB Edition      | State        | Enabled | Flag                                                  | Remarks                                                                  |
|:--------------------------------------------------------------------------------------|:-----------------|:----------------------|:-----------|:-----------------|:----------------------|:-------------|:--------|:------------------------------------------------------|:-------------------------------------------------------------------------|
| [Operator Ephemeral Volumes](/docs/design/features/ephemeral_volumes.md)              | 1.2.31           | Community, Enterprise | 1.2.2      | >= 3.8.0         | Community, Enterprise | Beta         | False   | --deployment.feature.ephemeral-volumes                | N/A                                                                      |
| [Rebalancer V2](/docs/design/features/rebalancer_v2.md)                               | 1.2.31           | Community, Enterprise | 1.2.31     | >= 3.10.0        | Community, Enterprise | Alpha        | False   | --deployment.feature.rebalancer-v2                    | N/A                                                                      |
| [Secured containers](/docs/design/features/secured_containers.md)                     | 1.2.31           | Community, Enterprise | 1.2.31     | >= 3.8.0         | Community, Enterprise | Alpha        | False   | --deployment.feature.secured-containers               | If set to True Operator will run containers in secure mode               |
| Version Check V2                                                                      | 1.2.31           | Community, Enterprise | 1.2.31     | >= 3.8.0         | Community, Enterprise | Alpha        | False   | --deployment.feature.upgrade-version-check-V2         | N/A                                                                      |
| [Force Rebuild Out Synced Shards](/docs/design/features/rebuild_out_synced_shards.md) | 1.2.27           | Community, Enterprise | 1.2.27     | >= 3.8.0         | Community, Enterprise | Production   | False   | --deployment.feature.force-rebuild-out-synced-shards  | It should be used only if user is aware of the risks.                    |
| [Spec Default Restore](/docs/design/features/deployment_spec_defaults.md)             | 1.2.25           | Community, Enterprise | 1.2.21     | >= 3.8.0         | Community, Enterprise | Beta         | True    | --deployment.feature.deployment-spec-defaults-restore | If set to False Operator will not change ArangoDeployment Spec           |
| Version Check                                                                         | 1.2.23           | Community, Enterprise | 1.1.4      | >= 3.8.0         | Community, Enterprise | Production   | True    | --deployment.feature.upgrade-version-check            | N/A                                                                      |
| [Failover Leader service](/docs/design/features/failover_leader_service.md)           | 1.2.13           | Community, Enterprise | 1.2.13     | >= 3.8.0         | Community, Enterprise | Production   | False   | --deployment.feature.failover-leadership              | N/A                                                                      |
| Graceful Restart                                                                      | 1.2.5            | Community, Enterprise | 1.0.7      | >= 3.8.0         | Community, Enterprise | Production   | True    | ---deployment.feature.graceful-shutdown               | N/A                                                                      |
| Encryption Key Rotation Support                                                       | 1.2.0            | Community, Enterprise | 1.0.3      | >= 3.8.0         | Enterprise            | NotSupported | False   | --deployment.feature.encryption-rotation              | N/A                                                                      |
| Operator Internal Metrics Exporter                                                    | 1.2.0            | Community, Enterprise | 1.2.0      | >= 3.8.0         | Community, Enterprise | Production   | True    | --deployment.feature.metrics-exporter                 | N/A                                                                      |
| Operator Maintenance Management Support                                               | 1.2.0            | Community, Enterprise | 1.0.7      | >= 3.8.0         | Community, Enterprise | Production   | True    | --deployment.feature.maintenance                      | N/A                                                                      |
| Optional Graceful Restart                                                             | 1.2.0            | Community, Enterprise | 1.2.5      | >= 3.8.0         | Community, Enterprise | Production   | False   | --deployment.feature.optional-graceful-shutdown       | N/A                                                                      |
| JWT Rotation Support                                                                  | 1.1.0            | Community, Enterprise | 1.0.3      | >= 3.8.0         | Enterprise            | Production   | True    | --deployment.feature.jwt-rotation                     | N/A                                                                      |
| TLS Runtime Rotation Support                                                          | 1.1.0            | Community, Enterprise | 1.0.4      | >= 3.8.0         | Enterprise            | Production   | True    | --deployment.feature.tls-rotation                     | N/A                                                                      |
| Operator Single Mode                                                                  | 1.0.4            | Community, Enterprise | 1.0.4      | >= 3.8.0         | Community, Enterprise | Production   | False   | --mode.single                                         | Only 1 instance of Operator allowed in namespace when feature is enabled |
| TLS SNI Support                                                                       | 1.0.3            | Community, Enterprise | 1.0.3      | >= 3.8.0         | Enterprise            | Production   | True    | --deployment.feature.tls-sni                          | N/A                                                                      |
| Disabling of liveness probes                                                          | 0.3.11           | Community, Enterprise | 0.3.10     | >= 3.8.0         | Community, Enterprise | Production   | True    | N/A                                                   | N/A                                                                      |
| Pod Disruption Budgets                                                                | 0.3.11           | Community, Enterprise | 0.3.10     | >= 3.8.0         | Community, Enterprise | Production   | True    | N/A                                                   | N/A                                                                      |
| Prometheus Metrics Exporter                                                           | 0.3.11           | Community, Enterprise | 0.3.10     | >= 3.8.0         | Community, Enterprise | Production   | True    | N/A                                                   | Prometheus required                                                      |
| Sidecar Containers                                                                    | 0.3.11           | Community, Enterprise | 0.3.10     | >= 3.8.0         | Community, Enterprise | Production   | True    | N/A                                                   | N/A                                                                      |
| Volume Claim Templates                                                                | 0.3.11           | Community, Enterprise | 0.3.10     | >= 3.8.0         | Community, Enterprise | Production   | True    | N/A                                                   | N/A                                                                      |
| Volume Resizing                                                                       | 0.3.11           | Community, Enterprise | 0.3.10     | >= 3.8.0         | Community, Enterprise | Production   | True    | N/A                                                   | N/A                                                                      |

<!-- END(featuresTable) -->

## Operator Community Edition (CE)

Image: `arangodb/kube-arangodb:1.2.31`

### Installation of latest CE release using Kubectl

```bash
kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/1.2.31/manifests/arango-crd.yaml
kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/1.2.31/manifests/arango-deployment.yaml
# To use `ArangoLocalStorage`, also run
kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/1.2.31/manifests/arango-storage.yaml
# To use `ArangoDeploymentReplication`, also run
kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/1.2.31/manifests/arango-deployment-replication.yaml
```

This procedure can also be used for upgrades and will not harm any
running ArangoDB deployments.

### Installation of latest CE release using kustomize

Installation using [kustomize](https://kubernetes.io/docs/tasks/manage-kubernetes-objects/kustomization/) looks like installation from yaml files,
but user is allowed to modify namespace or resource names without yaml modifications.

IT is recommended to use kustomization instead of handcrafting namespace in yaml files - kustomization will replace not only resource namespaces,
but also namespace references in resources like ClusterRoleBinding.

Example kustomization file:
```
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namespace: my-custom-namespace

bases:
  - https://github.com/arangodb/kube-arangodb/manifests/kustomize/deployment/?ref=1.0.3
```

### Installation of latest CE release using Helm

Only use this procedure for a new install of the operator. See below for
upgrades.

```bash
# The following will install the operator for `ArangoDeployment` &
# `ArangoDeploymentReplication` resources.
helm install https://github.com/arangodb/kube-arangodb/releases/download/1.2.31/kube-arangodb-1.2.31.tgz
# To use `ArangoLocalStorage`, set field `operator.features.storage` to true
helm install https://github.com/arangodb/kube-arangodb/releases/download/1.2.31/kube-arangodb-1.2.31.tgz --set "operator.features.storage=true"
```

### Upgrading the operator using Helm

To upgrade the operator to the latest version with Helm, you have to
delete the previous deployment and then install the latest. **HOWEVER**:
You *must not delete* the deployment of the custom resource definitions
(CRDs), or your ArangoDB deployments will be deleted!

Therefore, you have to use `helm list` to find the deployments for the
operator (`kube-arangodb`) and of the storage operator
(`kube-arangodb-storage`) and use `helm delete` to delete them using the
automatically generated deployment names. Here is an example of a `helm
list` output:

```
% helm list
NAME            	REVISION	UPDATED                 	STATUS  	CHART                               	APP VERSION	NAMESPACE
vetoed-ladybird 	1       	Mon Apr  8 11:36:58 2019	DEPLOYED	kube-arangodb-0.3.10-preview        	           	default  
```

So here, you would have to do

```bash
helm delete vetoed-ladybird
```

but **not delete `steely-mule`**. Then you could install the new version
with `helm install` as normal:

```bash
# The following will install the operator for `ArangoDeployment` &
# `ArangoDeploymentReplication` resources.
helm install https://github.com/arangodb/kube-arangodb/releases/download/1.2.31/kube-arangodb-1.2.31.tgz
# To use `ArangoLocalStorage`, set field `operator.features.storage` to true
helm install https://github.com/arangodb/kube-arangodb/releases/download/1.2.31/kube-arangodb-1.2.31.tgz --set "operator.features.storage=true"
```

## Operator Enterprise Edition (EE)

Image: `arangodb/kube-arangodb-enterprise:1.2.31`

### Installation of latest EE release using Kubectl

```bash
kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/1.2.31/manifests/enterprise-crd.yaml
kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/1.2.31/manifests/enterprise-deployment.yaml
# To use `ArangoLocalStorage`, also run
kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/1.2.31/manifests/enterprise-storage.yaml
# To use `ArangoDeploymentReplication`, also run
kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/1.2.31/manifests/enterprise-deployment-replication.yaml
```

This procedure can also be used for upgrades and will not harm any
running ArangoDB deployments.

### Installation of latest EE release using kustomize

Installation using [kustomize](https://kubernetes.io/docs/tasks/manage-kubernetes-objects/kustomization/) looks like installation from yaml files,
but user is allowed to modify namespace or resource names without yaml modifications.

IT is recommended to use kustomization instead of handcrafting namespace in yaml files - kustomization will replace not only resource namespaces,
but also namespace references in resources like ClusterRoleBinding.

Example kustomization file:
```
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namespace: my-custom-namespace

bases:
  - https://github.com/arangodb/kube-arangodb/manifests/kustomize-enterprise/deployment/?ref=1.0.3
```

### Installation of latest EE release using Helm

Only use this procedure for a new install of the operator. See below for
upgrades.

```bash
# The following will install the operator for `ArangoDeployment` &
# `ArangoDeploymentReplication` resources.
helm install https://github.com/arangodb/kube-arangodb/releases/download/1.2.31/kube-arangodb-1.2.31.tgz --set "operator.image=arangodb/kube-arangodb-enterprise:1.2.31"
# To use `ArangoLocalStorage`, set field `operator.features.storage` to true
helm install https://github.com/arangodb/kube-arangodb/releases/download/1.2.31/kube-arangodb-1.2.31.tgz --set "operator.image=arangodb/kube-arangodb-enterprise:1.2.31" --set "operator.features.storage=true"
```

### Upgrading the operator using Helm

To upgrade the operator to the latest version with Helm, you have to
delete the previous deployment and then install the latest. **HOWEVER**:
You *must not delete* the deployment of the custom resource definitions
(CRDs), or your ArangoDB deployments will be deleted!

Therefore, you have to use `helm list` to find the deployments for the
operator (`kube-arangodb`) and of the storage operator
(`kube-arangodb-storage`) and use `helm delete` to delete them using the
automatically generated deployment names. Here is an example of a `helm
list` output:

```
% helm list
NAME            	REVISION	UPDATED                 	STATUS  	CHART                               	APP VERSION	NAMESPACE
vetoed-ladybird 	1       	Mon Apr  8 11:36:58 2019	DEPLOYED	kube-arangodb-0.3.10-preview        	           	default  
```

So here, you would have to do

```bash
helm delete vetoed-ladybird
```

but **not delete `steely-mule`**. Then you could install the new version
with `helm install` as normal:

```bash
# The following will install the operator for `ArangoDeployment` &
# `ArangoDeploymentReplication` resources.
helm install https://github.com/arangodb/kube-arangodb/releases/download/1.2.31/kube-arangodb-1.2.31.tgz --set "operator.image=arangodb/kube-arangodb-enterprise:1.2.31"
# To use `ArangoLocalStorage`, set field `operator.features.storage` to true
helm install https://github.com/arangodb/kube-arangodb/releases/download/1.2.31/kube-arangodb-1.2.31.tgz --set "operator.image=arangodb/kube-arangodb-enterprise:1.2.31" --set "operator.features.storage=true"
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
