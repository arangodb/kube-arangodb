# ArangoDB Kubernetes Operator

"Starter for Kubernetes"

State: In heavy development. DO NOT USE FOR ANY PRODUCTION LIKE PURPOSE! THINGS WILL CHANGE.

- [User manual](./docs/user/README.md)
- [Design documents](./docs/design/README.md)

## Building

```bash
DOCKERNAMESPACE=<your dockerhub account> make
kubectl apply -f manifests/crd.yaml
kubectl apply -f manifests/arango-deployment-dev.yaml
# To use `ArangoLocalStorage`, also run
kubectl apply -f manifests/arango-storage-dev.yaml
```
