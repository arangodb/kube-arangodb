---
layout: page
title: Google Cloud Storage
parent: Storage
grand_parent: ArangoDBPlatform
nav_order: 2
---

# Integration

In order to connect to the GCS (Google Cloud Storage):

## GCP ServiceAccount

ServiceAccount with access to the storage needs to be saved in the secret.

```shell
kubectl create secret generic credentials --from-file 'serviceAccount=<ServiceAccount JSON File>'
```

## Permissions

The ServiceAccount needs the following IAM permissions on the configured bucket:

| Permission | Purpose |
|---|---|
| `storage.buckets.get` | Check bucket existence |
| `storage.buckets.create` | Create the bucket if it does not exist |
| `storage.objects.get` | Read objects and object metadata |
| `storage.objects.create` | Write objects |
| `storage.objects.list` | List objects |
| `storage.objects.delete` | Delete objects |

The predefined role `roles/storage.admin` grants all of the above. If the bucket already exists and is
managed externally, `roles/storage.objectAdmin` together with `storage.buckets.get` is sufficient (and
`storage.buckets.create` can be omitted).

## Object

Once the Secret is created, we are able to create ArangoPlatformStorage.

```
echo "---
apiVersion: platform.arangodb.com/v1beta1
kind: ArangoPlatformStorage
metadata:
  name: deployment
  namespace: namespace
spec:
  backend:
    gcs:
      bucketName: <Bucket Name>
      bucketPath: <Bucket Path>
      credentialsSecret:
        name: credentials
      projectID: <Google Project ID>
" | kubectl apply -f -
```