---
layout: page
title: MinIO
parent: Storage
grand_parent: ArangoDBPlatform
nav_order: 4
---

# Integration

In order to connect to the MinIO, or any other S3 Compatible storage in the ArangoPlatform:

## MinIO Access Keys

Storage Integration requires static credentials in order to access MinIO API. Credentials can be provided via the Kubernetes Secret.

```shell
kubectl create secret generic credentials --from-literal 'accessKey=<MinIO Access Key ID>' --from-literal 'secretKey=<MinIO Secret Access Key>'
```

## MinIO TLS Certificate

```shell
kubectl create secret generic ca --from-file 'ca.crt=<Certificate Path>'
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
    s3:
      bucketName: <Bucket Name>
      bucketPath: <Bucket Path>
      credentialsSecret:
        name: credentials
      caSecret:
        name: ca
      endpoint: https://minio.namespace.svc # Minio Endpoint
" | kubectl apply -f -
```
