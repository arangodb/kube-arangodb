---
layout: page
title: Azure Blob Storage
parent: Storage
grand_parent: ArangoDBPlatform
nav_order: 3
---

# Integration

In order to connect to the Azure Blob storage:

## Azure Credentials

Client ID & Secret with access to the storage container and accounts need to be saved in the secret.

```shell
kubectl create secret generic credentials --from-literal 'clientId=<Azure Client ID>' --from-literal 'clientSecret=<Azure Client Secret>'
```

## Permissions

The provided Client ID needs read/write access to the blobs in the configured container. Assign the
built-in Azure RBAC role **Storage Blob Data Contributor**, scoped to the storage account or the
container:

| Operation | Purpose |
|---|---|
| Read container properties | Check container existence |
| Read blob / blob metadata | Read objects |
| Write blob | Write objects |
| List blobs | List objects |
| Delete blob | Delete objects |

The **Storage Blob Data Contributor** role covers all of the above. The container must already exist -
the integration does not create it.

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
    azureBlobStorage:
      bucketName: <Bucket Name>
      bucketPath: <Bucket Path>
      credentialsSecret:
        name: credentials
      tenantID: <Azure Tenant ID>
      accountName: <Azure Storage Account Name>
      endpoint: <Azure Storage Endpoint in case of Private Connection>
" | kubectl apply -f -
```