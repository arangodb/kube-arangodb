# ArangoDB Kubernetes Operator

[![Docker Pulls](https://img.shields.io/docker/pulls/arangodb/kube-arangodb.svg)](https://hub.docker.com/r/arangodb/kube-arangodb/)

ArangoDB Kubernetes Operator helps to run ArangoDB deployments
on Kubernetes clusters.

To get started, follow the Installation instructions below and/or
read the [tutorial](https://www.arangodb.com/docs/stable/tutorials-kubernetes.html).

## State

The ArangoDB Kubernetes Operator is still in **development**.

Running ArangoDB deployments (single, active-failover or cluster)
is reasonably stable, and we're in the process of validating
production readiness of various Kubernetes platforms.

The feature set of the ArangoDB Kubernetes Operator is close to what
it is intended to be.

[Documentation](./docs/README.md)

### Production readiness state

Beginning with Version 0.3.11 we maintain a production readiness
state for individual new features, since we expect that new
features will first be released with an "alpha" or "beta" readiness
state and over time move to full "production readiness".

The following table has the general readiness state, the table below
covers individual newer features separately.

| Platform            | Kubernetes Version | ArangoDB Version | ArangoDB Operator Version | State       | Remarks               | Provider Remarks                   |
|---------------------|--------------------|------------------|---------------------------|-------------|-----------------------|------------------------------------|
| Google GKE          | 1.14               | >= 3.3.13        |                           | Production  | Don't use micro nodes |                                    |
| Google GKE          | 1.15               | >= 3.3.13        |                           | Production  | Don't use micro nodes |                                    |
| Azure AKS           | 1.14               | >= 3.3.13        |                           | Production  |                       |                                    |
| Azure AKS           | 1.15               | >= 3.3.13        |                           | Production  |                       |                                    |
| Amazon EKS          | 1.14               | >= 3.3.13        |                           | Production  |                       | [Amazon EKS](./docs/providers/eks) |
| IBM Cloud           | 1.14               | >= 3.4.6.1       | >= 0.3.11                 | Production  |                       |                                    |
| OpenShift           | 3.11               | >= 3.3.13        |                           | Production  |                       |                                    |
| OpenShift           | 4.2                | >= 3.3.13        |                           | In Progress |                       |                                    |
| BareMetal (kubeadm) | 1.14               | >= 3.3.13        |                           | Production  |                       |                                    |
| Minikube            | 1.14               | >= 3.3.13        |                           | Devel Only  |                       |                                    |
| Other               | 1.14               | >= 3.3.13        |                           | Devel Only  |                       |                                    |

Feature-wise production readiness table:

| Feature                      | ArangoDB K8s Operator Version         | Production Readiness      | Remarks           |
|------------------------------|---------------------------------------|---------------------------|-------------------|
| Pod Disruption Budgets       | 0.3.10                                | new - alpha               |                   |
|                              | 0.3.11                                | beta                      |                   |
| Volume Resizing              | 0.3.10                                | new - beta                |                   |
|                              | 0.3.11                                | beta                      |                   |
| Disabling of liveness probes | 0.3.10                                | new - beta                |                   |
|                              | 0.3.11                                | production ready          |                   |
| Volume Claim Templates       | 0.3.11                                | new - alpha               |                   |
| Prometheus Metrics export    | 0.3.11                                | new - alpha               | needs Prometheus  |
| User sidecar containers      | 0.3.11                                | new - alpha               |                   |

## Release notes for 0.3.16

In this release we have reworked the Helm charts. One notable change is
that we now create a new service account specifically for the operator.
The actual deployment still runs by default under the `default` service
account unless one changes that. Note that the service account under
which the ArangoDB runs needs a small set of extra permissions. For
the `default` service account we grant them when the operator is
deployed. If you use another service account you have to grant these
permissions yourself. See
[here](docs/Manual/Deployment/Kubernetes/DeploymentResource.md#specgroupserviceaccountname-string)
for details.

## Installation of latest release using Kubectl

```bash
kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/1.0.3/manifests/arango-crd.yaml
kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/1.0.3/manifests/arango-deployment.yaml
# To use `ArangoLocalStorage`, also run
kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/1.0.3/manifests/arango-storage.yaml
# To use `ArangoDeploymentReplication`, also run
kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/1.0.3/manifests/arango-deployment-replication.yaml
```

This procedure can also be used for upgrades and will not harm any
running ArangoDB deployments.

## Installation of latest release using kustomize

Installation using [kustomize](https://kubernetes.io/docs/tasks/manage-kubernetes-objects/kustomization/) looks like installation from yaml files,
but user is allowed to modify namespace or resource names without yaml modifications.

IT is recommended to use kustomization instead of handcrafting namespace in yaml files - kustomization will replace not only resource namespaces,
but also namespace references in resources like ClusterRoleBinding.

Example kustomization file:
```
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

bases:
  - https://github.com/arangodb/kube-arangodb/manifests/kustomize/crd/?ref=1.0.3
```

## Installation of latest release using Helm

Only use this procedure for a new install of the operator. See below for
upgrades.

```bash
# The following will install the custom resources required by the operators.
helm install https://github.com/arangodb/kube-arangodb/releases/download/1.0.3/kube-arangodb-crd-1.0.3.tgz
# The following will install the operator for `ArangoDeployment` &
# `ArangoDeploymentReplication` resources.
helm install https://github.com/arangodb/kube-arangodb/releases/download/1.0.3/kube-arangodb-1.0.3.tgz
# To use `ArangoLocalStorage`, set field `operator.features.storage` to true
helm install https://github.com/arangodb/kube-arangodb/releases/download/1.0.3/kube-arangodb-1.0.3.tgz --set "operator.features.storage=true"
```

## Upgrading the operator using Helm

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
steely-mule     	1       	Sun Mar 31 21:11:07 2019	DEPLOYED	kube-arangodb-crd-0.3.9             	           	default  
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
helm install https://github.com/arangodb/kube-arangodb/releases/download/1.0.3/kube-arangodb-1.0.3.tgz
# To use `ArangoLocalStorage`, set field `operator.features.storage` to true
helm install https://github.com/arangodb/kube-arangodb/releases/download/1.0.3/kube-arangodb-1.0.3.tgz --set "operator.features.storage=true"
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
