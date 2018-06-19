# ArangoDB Kubernetes Operator

ArangoDB Kubernetes Operator helps do run ArangoDB deployments
on Kubernetes clusters.

To get started, follow the Installation instructions below and/or
read the [tutorial](./docs/Manual/Tutorials/Kubernetes/README.md).

## State

The ArangoDB Kubernetes Operator is still in **heavy development**.

Running ArangoDB deployments (single, active-failover or cluster)
is becoming reasonably stable, but you should **not yet use it for production
environments**.

The feature set of the ArangoDB Kubernetes Operator is close to what
it is intended to be, with the exeption of Datacenter to Datacenter replication
support. That is still completely missing.

[Documentation](./docs/README.md)

## Installation of latest release

```bash
kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/0.2.1/manifests/crd.yaml
kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/0.2.1/manifests/arango-deployment.yaml
# To use `ArangoLocalStorage`, also run
kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/0.2.1/manifests/arango-storage.yaml
```

## Building

```bash
DOCKERNAMESPACE=<your dockerhub account> make
kubectl apply -f manifests/crd.yaml
kubectl apply -f manifests/arango-deployment-dev.yaml
# To use `ArangoLocalStorage`, also run
kubectl apply -f manifests/arango-storage-dev.yaml
```
