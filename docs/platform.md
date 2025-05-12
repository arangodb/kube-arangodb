---
layout: page
has_children: true
title: ArangoDBPlatform
has_toc: false
---

# Platform


#### Community & Enterprise Edition

[Full CustomResourceDefinition reference ->](./api/ArangoPlatformStorage.V1Alpha1.md)

This instruction covers only the steps to enable ArangoPlatform in Kubernetes cluster with already running ArangoDeployment.
If you don't have one yet, consider checking [kube-arangodb installation guide](./using-the-operator.md) and [ArangoDeployment CR description](./deployment-resource-reference.md).

### To enable Platform on your cluster, follow next steps:

1) Install [CertManager](https://github.com/cert-manager/cert-manager) on your cluster.
 
2) Enable Webhooks. e.g. if you are using Helm package, add `--set "webhooks.enabled=true"` option to the Helm command.

3) Enable Managed Certificates. e.g. if you are using Helm package, add `--set "certificate.enabled=true"` option to the Helm command.

4) Enable Gateways in the ArangoDeployment. Set `.spec.gateway.enabled` and `.spec.gateway.dynamic` to True
