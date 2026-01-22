---
layout: page
title: How to set a license key or license credentials
parent: How to ...
---

# How to set a license key

After deploying the ArangoDB Kubernetes operator, use the command below to deploy your [license key](https://docs.arango.ai/arangodb/stable/operations/administration/license-management/)
as a secret, which is required for the ArangoDB Enterprise Edition starting with version 3.9:

```bash
# For a license key
kubectl create secret generic arango-license-key --from-literal=token-v2="<license-string>"

# For license credentials (managed license, from v3.12.6 onward)
kubectl create secret generic arango-license-key --namespace="<namespace>" --from-literal=license-client-id="<license-client-id>" --from-literal=license-client-secret="<license-client-secret>"
```

Then reference the newly created secret in the `ArangoDeployment` specification:

```yaml
spec:
  # [...]
  license:
    secretName: arango-license-key
```

To update the license information, delete the secret and create a new one with the same name and the updated information:

```bash
kubectl delete secret --namespace arangodb arango-license-key
kubectl create secret generic arango-license-key ...
```
