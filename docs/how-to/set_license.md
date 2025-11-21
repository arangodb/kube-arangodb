---
layout: page
title: How to set a license key
parent: How to ...
---

# How to set a license key

After deploying the ArangoDB Kubernetes operator, use the command below to deploy your [license key](https://docs.arangodb.com/stable/operations/administration/license-management/)
as a secret which is required for the Enterprise Edition starting with version 3.9:

```bash
# For the License Key
kubectl create secret generic arango-license-key --from-literal=token-v2="<license-string>"

# For the License Manager Key
kubectl create secret generic arango-license-key --from-literal=license-client-id="<license-client-id>" --from-literal=license-client-secret="<license-client-secret>"
```


Then specify the newly created secret in the ArangoDeploymentSpec:
```yaml
spec:
  # [...]
  license:
    secretName: arango-license-key
```
