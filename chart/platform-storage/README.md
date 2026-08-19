# platform-storage

Deploys an in-cluster, S3-compatible object store ([MinIO](https://min.io/)) and, optionally, wires it
to the ArangoDB Platform by rendering an
[`ArangoPlatformStorage`](../../docs/api/ArangoPlatformStorage.V1Beta1.md) resource whose `s3` backend
points at that MinIO Service. See also the [MinIO storage integration](../../docs/platform/storage/minio.md)
docs.

It is intended for self-hosted / development environments that need object storage for the ArangoDB
Platform without an external S3, GCS or Azure Blob Storage account.

## Prerequisites

- The ArangoDB Kubernetes Operator installed with platform/storage support (the
  `platform.arangodb.com/v1beta1` `ArangoPlatformStorage` CRD present).
- A default `StorageClass`, or set `storage.class`.

## Installing

```sh
# MinIO only
helm install my-storage chart/platform-storage

# MinIO + an ArangoPlatformStorage named "my-storage" pointing at it
helm install my-storage chart/platform-storage --set deployment.name=my-storage
```

## What it deploys

| Resource | Name | Notes |
|---|---|---|
| Deployment | `<release>` | single MinIO replica (`server /data`) |
| Service | `<release>` | port `9000` |
| PersistentVolumeClaim | `<release>` | `storage.size` on `storage.class` |
| Secret | `<release>-root-credentials` | `MINIO_ROOT_USER` / `MINIO_ROOT_PASSWORD` (password generated on first install, reused on upgrade) |
| Secret | `<release>-credentials` | `accessKey` / `secretKey` consumed by the `ArangoPlatformStorage` |
| ArangoPlatformStorage | `<deployment.name>` | only when `deployment.name` is set; `s3` backend → `http://<release>.<namespace>.svc:9000` |

## Configuration

| Key | Default | Description |
|---|---|---|
| `deployment.name` | `""` | Name of the `ArangoPlatformStorage` to render. Empty deploys MinIO only. |
| `storage.class` | `""` | `StorageClass` for the MinIO PVC (empty = cluster default). |
| `storage.size` | `5Gi` | Size of the MinIO PVC. |
| `minio.image` | `minio/minio:latest` | MinIO container image. |
| `minio.imagePullPolicy` | `IfNotPresent` | Image pull policy. |
| `minio.imagePullSecrets` | `[]` | Image pull secrets. |
| `minio.resources` | `250m` / `512Mi` (requests & limits) | MinIO container resources. |
| `minio.nodeSelector` | `{}` | Node selector for the MinIO Pod. |
| `minio.tolerations` | `[]` | Tolerations for the MinIO Pod. |

See [`values.yaml`](values.yaml) for the full set of defaults.
