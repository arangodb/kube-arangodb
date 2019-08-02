# ArangoDB Kubernetes Operator

[![Docker Pulls](https://img.shields.io/docker/pulls/arangodb/kube-arangodb.svg)](https://hub.docker.com/r/arangodb/kube-arangodb/)

ArangoDB Kubernetes Operator helps to run ArangoDB deployments
on Kubernetes clusters.

To get started, follow the Installation instructions below and/or
read the [tutorial](./docs/Manual/Tutorials/Kubernetes/README.md).

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

| Platform             | Kubernetes version | ArangoDB version | ArangoDB K8s Operator Version | State | Production ready | Remarks               |
|----------------------|--------------------|------------------|-------------------------------|-------|------------------|-----------------------|
| Google GKE           | 1.10               | >= 3.3.13        |                               | Runs  | Yes              | Don't use micro nodes |
| Google GKE           | 1.11               | >= 3.3.13        |                               | Runs  | Yes              | Don't use micro nodes |
| Amazon EKS           | 1.11               | >= 3.3.13        |                               | Runs  | Yes              |                       |
| Pivotal PKS          | 1.11               | >= 3.3.13        |                               | Runs  | Yes              |                       |
| IBM Cloud            | 1.11               | >= 3.4.5         |          >= 0.3.11            | Runs  | Yes              |                       |
| IBM Cloud            | 1.12               | >= 3.4.5         |          >= 0.3.11            | Runs  | Yes              |                       |
| IBM Cloud            | 1.13               | >= 3.4.6.1       |          >= 0.3.11            | Runs  | Yes              |                       |
| Amazon & Kops        | 1.10               | >= 3.3.13        |                               | Runs  | No               |                       |
| Azure AKS            | 1.10               | >= 3.3.13        |                               | Runs  | No               |                       |
| OpenShift            | 1.10               | >= 3.3.13        |                               | Runs  | No               |                       |
| Bare metal (kubeadm) | 1.10               | >= 3.3.13        |                               | Runs  | Yes              |                       |
| Bare metal (kubeadm) | 1.11               | >= 3.3.13        |                               | Runs  | Yes              |                       |
| Bare metal (kubeadm) | 1.12               | >= 3.3.13        |                               | Runs  | In progress      |                       |
| Bare metal (kubeadm) | 1.13               | >= 3.3.13        |                               | Runs  | Yes              |                       |
| Bare metal (kubeadm) | 1.14               | >= 3.3.13        |                               | Runs  | In progress      |                       |
| Minikube             | 1.10               | >= 3.3.13        |                               | Runs  | Not intended     |                       |
| Docker for Mac Edge  | 1.10               | >= 3.3.13        |                               | Runs  | Not intended     |                       |
| Scaleway Kubernetes  | 1.10               | >= 3.3.13        | ?                             | No    |                  |                       |

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

## Installation of latest release using Kubectl

```bash
kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/0.3.13/manifests/arango-crd.yaml
kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/0.3.13/manifests/arango-deployment.yaml
# To use `ArangoLocalStorage`, also run
kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/0.3.13/manifests/arango-storage.yaml
# To use `ArangoDeploymentReplication`, also run
kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/0.3.13/manifests/arango-deployment-replication.yaml
```

This procedure can also be used for upgrades and will not harm any
running ArangoDB deployments.

## Installation of latest release using Helm

Only use this procedure for a new install of the operator. See below for
upgrades.

```bash
# The following will install the custom resources required by the operators.
helm install https://github.com/arangodb/kube-arangodb/releases/download/0.3.13/kube-arangodb-crd.tgz
# The following will install the operator for `ArangoDeployment` &
# `ArangoDeploymentReplication` resources.
helm install https://github.com/arangodb/kube-arangodb/releases/download/0.3.13/kube-arangodb.tgz
# To use `ArangoLocalStorage`, also run
helm install https://github.com/arangodb/kube-arangodb/releases/download/0.3.13/kube-arangodb-storage.tgz
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
intent-camel    	1       	Mon Apr  8 11:37:52 2019	DEPLOYED	kube-arangodb-storage-0.3.10-preview	           	default  
steely-mule     	1       	Sun Mar 31 21:11:07 2019	DEPLOYED	kube-arangodb-crd-0.3.9             	           	default  
vetoed-ladybird 	1       	Mon Apr  8 11:36:58 2019	DEPLOYED	kube-arangodb-0.3.10-preview        	           	default  
```

So here, you would have to do

```bash
helm delete intent-camel
helm delete vetoed-ladybird
```

but **not delete `steely-mule`**. Then you could install the new version
with `helm install` as normal:

```bash
# The following will install the operator for `ArangoDeployment` &
# `ArangoDeploymentReplication` resources.
helm install https://github.com/arangodb/kube-arangodb/releases/download/0.3.13/kube-arangodb.tgz
# To use `ArangoLocalStorage`, also run
helm install https://github.com/arangodb/kube-arangodb/releases/download/0.3.13/kube-arangodb-storage.tgz
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
