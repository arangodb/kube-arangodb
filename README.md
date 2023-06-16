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

<!-- START(metricsTable) -->
| Platform            | Kubernetes Version | ArangoDB Version | State      | Remarks                                   | Provider Remarks                   |
|:--------------------|:-------------------|:-----------------|:-----------|:------------------------------------------|:-----------------------------------|
| Google GKE          | 1.21-1.25          | >= 3.6.0         | Production | Don't use micro nodes                     |                                    |
| Azure AKS           | 1.21-1.24          | >= 3.6.0         | Production |                                           |                                    |
| Amazon EKS          | 1.21-1.24          | >= 3.6.0         | Production |                                           | [Amazon EKS](./docs/providers/eks) |
| IBM Cloud           | 1.17               | >= 3.6.0         | Deprecated | Support will be dropped in Operator 1.5.0 |                                    |
| IBM Cloud           | 1.18-1.21          | >= 3.6.0         | Production |                                           |                                    |
| OpenShift           | 3.11               | >= 3.6.0         | Deprecated | Support will be dropped in Operator 1.5.0 |                                    |
| OpenShift           | 4.2-4.11           | >= 3.6.0         | Production |                                           |                                    |
| BareMetal (kubeadm) | <= 1.20            | >= 3.6.0         | Deprecated | Support will be dropped in Operator 1.5.0 |                                    |
| BareMetal (kubeadm) | 1.21-1.25          | >= 3.6.0         | Production |                                           |                                    |
| Minikube            | 1.21-1.25          | >= 3.6.0         | Devel Only |                                           |                                    |
| Other               | 1.21-1.25          | >= 3.6.0         | Devel Only |                                           |                                    |

<!-- END(metricsTable) -->

Feature-wise production readiness table:

| Feature                                 | Operator Version | ArangoDB Version | ArangoDB Edition      | Introduced | State        | Enabled | Flag                                                  | Remarks                                                                  |
|-----------------------------------------|------------------|------------------|-----------------------|------------|--------------|---------|-------------------------------------------------------|--------------------------------------------------------------------------|
| Pod Disruption Budgets                  | 0.3.11           | Any              | Community, Enterprise | 0.3.10     | Production   | True    | N/A                                                   | N/A                                                                      |
| Volume Resizing                         | 0.3.11           | Any              | Community, Enterprise | 0.3.10     | Production   | True    | N/A                                                   | N/A                                                                      |
| Disabling of liveness probes            | 0.3.11           | Any              | Community, Enterprise | 0.3.10     | Production   | True    | N/A                                                   | N/A                                                                      |
| Volume Claim Templates                  | 1.0.0            | Any              | Community, Enterprise | 0.3.10     | Production   | True    | N/A                                                   | N/A                                                                      |
| Prometheus Metrics Exporter             | 1.0.0            | Any              | Community, Enterprise | 0.3.10     | Production   | True    | N/A                                                   | Prometheus required                                                      |
| Sidecar Containers                      | 1.0.0            | Any              | Community, Enterprise | 0.3.10     | Production   | True    | N/A                                                   | N/A                                                                      |
| Operator Single Mode                    | 1.0.4            | Any              | Community, Enterprise | 1.0.4      | Production   | False   | --mode.single                                         | Only 1 instance of Operator allowed in namespace when feature is enabled |
| TLS SNI Support                         | 1.0.3            | >= 3.7.0         | Enterprise            | 1.0.3      | Production   | True    | --deployment.feature.tls-sni                          | N/A                                                                      |
| TLS Runtime Rotation Support            | 1.1.0            | > 3.7.0          | Enterprise            | 1.0.4      | Production   | True    | --deployment.feature.tls-rotation                     | N/A                                                                      |
| JWT Rotation Support                    | 1.1.0            | > 3.7.0          | Enterprise            | 1.0.3      | Production   | True    | --deployment.feature.jwt-rotation                     | N/A                                                                      |
| Encryption Key Rotation Support         | 1.2.0            | > 3.7.0          | Enterprise            | 1.0.3      | NotSupported | False   | --deployment.feature.encryption-rotation              | N/A                                                                      |
| Version Check                           | 1.1.4            | >= 3.6.0         | Community, Enterprise | 1.1.4      | Alpha        | False   | --deployment.feature.upgrade-version-check            | N/A                                                                      |
| Version Check                           | 1.2.23           | >= 3.6.0         | Community, Enterprise | 1.1.4      | Production   | True    | --deployment.feature.upgrade-version-check            | N/A                                                                      |
| Operator Maintenance Management Support | 1.2.0            | >= 3.6.0         | Community, Enterprise | 1.0.7      | Production   | True    | --deployment.feature.maintenance                      | N/A                                                                      |
| Graceful Restart                        | 1.2.5            | >= 3.6.0         | Community, Enterprise | 1.0.7      | Production   | True    | --deployment.feature.graceful-shutdown                | N/A                                                                      |
| Optional Graceful Restart               | 1.2.25           | >= 3.6.0         | Community, Enterprise | 1.2.5      | Beta         | True    | --deployment.feature.optional-graceful-shutdown       | N/A                                                                      |
| Operator Internal Metrics Exporter      | 1.2.0            | >= 3.6.0         | Community, Enterprise | 1.2.0      | Production   | True    | --deployment.feature.metrics-exporter                 | N/A                                                                      |
| Operator Ephemeral Volumes              | 1.2.2            | >= 3.7.0         | Community, Enterprise | 1.2.2      | Alpha        | False   | --deployment.feature.ephemeral-volumes                | N/A                                                                      |
| Spec Default Restore                    | 1.2.21           | >= 3.7.0         | Community, Enterprise | 1.2.21     | Beta         | True    | --deployment.feature.deployment-spec-defaults-restore | If set to False Operator will not change ArangoDeployment Spec           |
| Force Rebuild Out Synced Shards         | 1.2.27           | >= 3.8.0         | Community, Enterprise | 1.2.27     | Beta         | False   | --deployment.feature.force-rebuild-out-synced-shards  | It should be used only if user is aware of the risks.                    |

## Operator Community Edition (CE)

Image: `arangodb/kube-arangodb:1.2.30`

### Installation of latest CE release using Kubectl

```bash
kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/1.2.30/manifests/arango-crd.yaml
kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/1.2.30/manifests/arango-deployment.yaml
# To use `ArangoLocalStorage`, also run
kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/1.2.30/manifests/arango-storage.yaml
# To use `ArangoDeploymentReplication`, also run
kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/1.2.30/manifests/arango-deployment-replication.yaml
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
helm install https://github.com/arangodb/kube-arangodb/releases/download/1.2.30/kube-arangodb-1.2.30.tgz
# To use `ArangoLocalStorage`, set field `operator.features.storage` to true
helm install https://github.com/arangodb/kube-arangodb/releases/download/1.2.30/kube-arangodb-1.2.30.tgz --set "operator.features.storage=true"
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
helm install https://github.com/arangodb/kube-arangodb/releases/download/1.2.30/kube-arangodb-1.2.30.tgz
# To use `ArangoLocalStorage`, set field `operator.features.storage` to true
helm install https://github.com/arangodb/kube-arangodb/releases/download/1.2.30/kube-arangodb-1.2.30.tgz --set "operator.features.storage=true"
```

## Operator Enterprise Edition (EE)

Image: `arangodb/kube-arangodb-enterprise:1.2.30`

### Installation of latest EE release using Kubectl

```bash
kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/1.2.30/manifests/enterprise-crd.yaml
kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/1.2.30/manifests/enterprise-deployment.yaml
# To use `ArangoLocalStorage`, also run
kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/1.2.30/manifests/enterprise-storage.yaml
# To use `ArangoDeploymentReplication`, also run
kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/1.2.30/manifests/enterprise-deployment-replication.yaml
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
helm install https://github.com/arangodb/kube-arangodb/releases/download/1.2.30/kube-arangodb-1.2.30.tgz --set "operator.image=arangodb/kube-arangodb-enterprise:1.2.30"
# To use `ArangoLocalStorage`, set field `operator.features.storage` to true
helm install https://github.com/arangodb/kube-arangodb/releases/download/1.2.30/kube-arangodb-1.2.30.tgz --set "operator.image=arangodb/kube-arangodb-enterprise:1.2.30" --set "operator.features.storage=true"
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
helm install https://github.com/arangodb/kube-arangodb/releases/download/1.2.30/kube-arangodb-1.2.30.tgz --set "operator.image=arangodb/kube-arangodb-enterprise:1.2.30"
# To use `ArangoLocalStorage`, set field `operator.features.storage` to true
helm install https://github.com/arangodb/kube-arangodb/releases/download/1.2.30/kube-arangodb-1.2.30.tgz --set "operator.image=arangodb/kube-arangodb-enterprise:1.2.30" --set "operator.features.storage=true"
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
