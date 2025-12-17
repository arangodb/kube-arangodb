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
      projectID: gcr-for-testing
" | kubectl apply -f -
```