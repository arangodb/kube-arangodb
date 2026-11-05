# platform-storage

Deploys an in-cluster, S3-compatible object store — [MinIO](https://min.io/) (default) or a single-pod
[SeaweedFS](https://github.com/seaweedfs/seaweedfs), selected via `backend` — and, optionally, wires it
to the ArangoDB Platform by rendering an
[`ArangoPlatformStorage`](../../docs/api/ArangoPlatformStorage.V1Beta1.md) resource whose `s3` backend
points at that Service. See also the [MinIO](../../docs/platform/storage/minio.md) and
[SeaweedFS](../../docs/platform/storage/seaweedfs.md) storage integration docs.

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

# SeaweedFS instead of MinIO
helm install my-storage chart/platform-storage --set backend=seaweedfs --set deployment.name=my-storage
```

## What it deploys

| Resource | Name | Notes |
|---|---|---|
| Deployment | `<release>` | single replica — MinIO (`server /data`) or SeaweedFS (`weed server -s3`) depending on `backend` |
| Service | `<release>` | port `9000` (minio) or `8333` (seaweedfs) |
| PersistentVolumeClaim | `<release>` | `storage.size` on `storage.class` |
| Secret | `<release>-root-credentials` | minio only: `MINIO_ROOT_USER` / `MINIO_ROOT_PASSWORD` (generated on first install, reused on upgrade) |
| Secret | `<release>-config` | seaweedfs only: `s3.json` identities consumed by `weed -s3.config` |
| Secret | `<release>-credentials` | `accessKey` / `secretKey` consumed by the `ArangoPlatformStorage` |
| ArangoPlatformStorage | `<deployment.name>` | only when `deployment.name` is set; `s3` backend → `http://<release>.<namespace>.svc:<9000\|8333>` |

## Configuration

| Key | Default | Description |
|---|---|---|
| `backend` | `minio` | S3 implementation to deploy: `minio` or `seaweedfs`. |
| `deployment.name` | `""` | Name of the `ArangoPlatformStorage` to render. Empty deploys the object store only. |
| `storage.class` | `""` | `StorageClass` for the data PVC (empty = cluster default). |
| `storage.size` | `5Gi` | Size of the data PVC. |
| `minio.image` | `minio/minio:latest` | MinIO container image (used when `backend: minio`). |
| `minio.imagePullPolicy` / `imagePullSecrets` / `resources` / `nodeSelector` / `tolerations` | see [`values.yaml`](values.yaml) | MinIO Pod settings. |
| `seaweedfs.image` | `chrislusf/seaweedfs:4.47` | SeaweedFS container image (used when `backend: seaweedfs`). |
| `seaweedfs.imagePullPolicy` / `imagePullSecrets` / `resources` / `nodeSelector` / `tolerations` | see [`values.yaml`](values.yaml) | SeaweedFS Pod settings. |

See [`values.yaml`](values.yaml) for the full set of defaults.
