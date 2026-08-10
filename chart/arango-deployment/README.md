# ArangoDeployment

Renders a single `database.arangodb.com/v1` **ArangoDeployment** (Single, Cluster or ActiveFailover)
managed by the [ArangoDB Kubernetes Operator](https://github.com/arangodb/kube-arangodb).

Core fields are exposed as schema-validated values; the raw `spec` value is deep-merged over the
generated spec so every field of the ArangoDeployment spec stays reachable.

## Prerequisites

The operator and the ArangoDeployment CRD must already be installed (charts `kube-arangodb` and
`kube-arangodb-crd`). This chart only creates the ArangoDeployment resource.

## Installing

```bash
helm install my-db ./chart/arango-deployment --namespace my-namespace
```

Single server:

```bash
helm install my-db ./chart/arango-deployment --set mode=Single
```

Per-group settings (each group maps 1:1 to `ServerGroupSpec`), and the raw `spec` escape hatch for
anything not surfaced explicitly:

```yaml
dbservers:
  count: 5
spec:
  tls:
    caSecretName: None
```

## Values

| Key | Description | Default |
|-----|-------------|---------|
| `nameOverride` | Overrides the ArangoDeployment name | `""` (release name) |
| `mode` | `Single` / `Cluster` / `ActiveFailover` | `Cluster` |
| `image` | ArangoDB server image | `arangodb/arangodb:latest` |
| `agents` / `dbservers` / `coordinators` / `single` | Per-group `ServerGroupSpec` (count, ...) | counts `3/3/3/1` |
| `spec` | Raw ArangoDeployment spec, deep-merged over the generated spec | `{}` |

Only the groups relevant to the selected `mode` are rendered:

- `Cluster` → `agents`, `dbservers`, `coordinators`
- `ActiveFailover` → `agents`, `single`
- `Single` → `single`

## Container Images

| Image | Override path |
|-------|---------------|
| `arangodb/arangodb:latest` | `image` |
