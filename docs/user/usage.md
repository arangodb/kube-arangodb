# Using the ArangoDB operator

## Installation

The ArangoDB operator needs to be installed in your Kubernetes
cluster first. To do so, clone this repository and run:

```bash
kubectl create -f manifests/arango-operator.yaml
```

## Cluster creation

Once the operator is running, you can create your ArangoDB cluster
by creating a custom resource and deploying it.

For example:

```bash
kubectl create -f examples/simple-cluster.yaml
```

## Cluster removal

To remove an existing cluster, delete the custom
resource. The operator will then delete all created resources.

For example:

```bash
kubectl delete -f examples/simple-cluster.yaml
```

## Operator removal

To remove the entire ArangoDB operator, remove all
clusters first and then remove the operator by running:

```bash
kubectl delete -f manifests/arango-operator.yaml
```
