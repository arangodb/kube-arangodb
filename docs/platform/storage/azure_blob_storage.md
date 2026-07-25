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

Two authentication methods are supported. Define **exactly one** of `credentialsSecret` (client
secret) or `clientCertificateSecret` (client certificate) on the backend.

### Client secret (service principal)

Store the Client ID and Client Secret with access to the storage container in a Secret:

```shell
kubectl create secret generic credentials \
  --from-literal 'clientId=<Azure Client ID>' \
  --from-literal 'clientSecret=<Azure Client Secret>'
```

### Client certificate

Store the certificate and its private key in a native `kubernetes.io/tls` Secret, then add the
Client ID as an extra key:

```shell
kubectl create secret tls credentials-cert --cert=tls.crt --key=tls.key
kubectl patch secret credentials-cert --type merge -p '{"stringData":{"clientId":"<Azure Client ID>"}}'
```

The Secret must contain `tls.crt`, `tls.key` and `clientId`. Both the certificate and the key are
required: the private key signs the token request and the certificate identifies the credential to
Azure by its thumbprint, so upload the same certificate's public key to the Azure app registration.

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

To use client-certificate authentication instead, replace `credentialsSecret` with
`clientCertificateSecret` (the two are mutually exclusive):

```yaml
spec:
  backend:
    azureBlobStorage:
      bucketName: <Bucket Name>
      bucketPath: <Bucket Path>
      clientCertificateSecret:
        name: credentials-cert
      tenantID: <Azure Tenant ID>
      accountName: <Azure Storage Account Name>
      endpoint: <Azure Storage Endpoint in case of Private Connection>
```