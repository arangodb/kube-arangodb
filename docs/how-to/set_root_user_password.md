---
layout: page
title: How to set root user password
parent: How to ...
---

# How to set root user password

1) Create a kubernetes [Secret](https://kubernetes.io/docs/tasks/configmap-secret/managing-secret-using-kubectl/) with root password:
```bash
kubectl create secret generic arango-root-pwd --from-literal=username=root --from-literal=password=<paste_your_password_here>
```

1) Then specify the newly created secret in the ArangoDeploymentSpec:
```yaml
spec:
  bootstrap:
    passwordSecretNames:
      root: arango-root-pwd
```