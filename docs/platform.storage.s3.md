---
layout: page
title: AWS S3
parent: Storage
nav_order: 1
---

# Integration

In order to connect to the AWS S3 storage in the ArangoPlatform:

## AWS S3 Access Keys

Storage Integration requires static credentials in order to access AWS S3 API. Credentials can be provided via the Kubernetes Secret.

```shell
kubectl create secret generic credentials --from-literal 'accessKey=<AWS Access Key ID>' --from-literal 'secretKey=<AWS Secret Access Key'
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
      allowInsecure: true # In case of public certs are not installed needs to be set to false
      bucketName: <Bucket Name>
      bucketPath: <Bucket Path>
      credentialsSecret:
        name: credentials 
      endpoint: https://s3.eu-central-1.amazonaws.com # AWS S3 Region Endpoint
      region: eu-central-1 # AWS Region
" | kubectl apply -f -
```