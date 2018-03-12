# ArangoDB Kubernetes Operator

"Starter for Kubernetes"

State: In development

- [User manual](./docs/user/README.md)
- [Design documents](./docs/design/README.md)

## Building

```bash
DOCKERNAMESPACE=<your dockerhub account> make
kubectl apply -f manifests/arango-operator-dev.yaml
```
