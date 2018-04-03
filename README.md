# ArangoDB Kubernetes Operator

"Starter for Kubernetes"

State: In heavy development. DO NOT USE FOR ANY PRODUCTION LIKE PURPOSE! THINGS WILL CHANGE.

- [Getting Started](./docs/Manual/GettingStarted/Kubernetes/README.md)
- [User manual](./docs/Manual/Deployment/Kubernetes/README.md)
- [Design documents](./docs/design/README.md)

## Installation of latest release

```bash
kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/0.0.1/manifests/crd.yaml
kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/0.0.1/manifests/arango-deployment.yaml
# To use `ArangoLocalStorage`, also run
kubectl apply -f https://raw.githubusercontent.com/arangodb/kube-arangodb/0.0.1/manifests/arango-storage.yaml
```

## Building

```bash
DOCKERNAMESPACE=<your dockerhub account> make
kubectl apply -f manifests/crd.yaml
kubectl apply -f manifests/arango-deployment-dev.yaml
# To use `ArangoLocalStorage`, also run
kubectl apply -f manifests/arango-storage-dev.yaml
```
