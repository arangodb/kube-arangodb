---
layout: page
title: SeaweedFS
parent: Storage
grand_parent: ArangoDBPlatform
nav_order: 5
---

# Integration

In order to connect to a [SeaweedFS](https://github.com/seaweedfs/seaweedfs), or any other S3 Compatible storage in the ArangoPlatform:

## Provided Helm Chart

The [`platform-storage`](../../../chart/platform-storage/README.md) chart deploys a single-pod SeaweedFS
and the `ArangoPlatformStorage` in one step:

```sh
helm upgrade -i platform-storage-integration ./chart/platform-storage/ --set backend=seaweedfs --set deployment.name=<ArangoDeployment Name>
```

## Generic

### SeaweedFS Access Keys

Storage Integration requires static credentials in order to access the SeaweedFS S3 API. Credentials can be provided via the Kubernetes Secret.

```shell
kubectl create secret generic credentials --from-literal 'accessKey=<SeaweedFS Access Key ID>' --from-literal 'secretKey=<SeaweedFS Secret Access Key>'
```

The SeaweedFS server itself is configured with matching identities via its `-s3.config` file (`s3.json`).

### Object

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
      allowInsecure: true
      endpoint: http://seaweedfs.namespace.svc:8333 # SeaweedFS S3 Endpoint
" | kubectl apply -f -
```
