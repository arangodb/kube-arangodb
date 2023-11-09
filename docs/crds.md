# Custom resources overview

Main CRDs:
- [ArangoDeployment](deployment-resource-reference.md)
- [ArangoDeploymentReplication](deployment-replication-resource-reference.md)
- [ArangoLocalStorage](storage-resource.md)
- [Backup](backup-resource.md)
- [BackupPolicy](backuppolicy-resource.md)

Operator manages the CustomResources based on CustomResourceDefinitions installed in your cluster.

There are different options how CustomResourceDefinitions can be created.

**Deprecated options:**
- Install CRDs directly from `manifests` folder.
- Install `kube-arangodb-crd` helm chart before installing `kube-arangodb` chart.
- Install CRDs using kustomize `all` or `crd` manifests.

**Recommended:**
Use `kube-arangodb` Helm chart. Chart itself does not contain CRDs.
Instead, operator will try to create the required CRDs on the first start.
Make sure that ServiceAccount for operator has permissions to `create` CustomResourceDefinitions.

To disable the automatic creation of CRDs, set `enableCRDManagement=false` operator command line option, e.g.:
```shell
helm install --generate-name https://github.com/arangodb/kube-arangodb/releases/download/$VER/kube-arangodb-$VER.tgz --set "operator.enableCRDManagement=false"
```

## Schema validation

Starting with v1.2.36, the [schema validation](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/#validation) is supported for all CRDs.

Schema validation can be enabled only on cluster with no CRDs installed or by upgrading your CR from one CRD version to another.

To enable creation of CRD with validation schema, pass additional args to operator command line, e.g.:
```
--crd.validation-schema=arangobackuppolicies.backup.arangodb.com=true --crd.validation-schema=arangodeployments.database.arangodb.com=false
```
