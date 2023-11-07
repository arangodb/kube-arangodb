# Using the ArangoDB Kubernetes Operator with Helm

[`Helm`](https://www.helm.sh/) is a package manager for Kubernetes, which enables
you to install various packages (include the ArangoDB Kubernetes Operator)
into your Kubernetes cluster.

The benefit of `helm` (in the context of the ArangoDB Kubernetes Operator)
is that it allows for a lot of flexibility in how you install the operator.
For example you can install the operator in a namespace other than
`default`.

## Charts

The ArangoDB Kubernetes Operator is contained in `helm` chart `kube-arangodb` which contains the operator for the
`ArangoDeployment`, `ArangoLocalStorage` and `ArangoDeploymentReplication` resource types.

## Configurable values for ArangoDB Kubernetes Operator

The following values can be configured when installing the
ArangoDB Kubernetes Operator with `helm`.

Values are passed to `helm` using an `--set=<key>=<value>` argument passed
to the `helm install` or `helm upgrade` command.

### `operator.image`

Image used for the ArangoDB Operator.

Default: `arangodb/kube-arangodb:latest`

### `operator.imagePullPolicy`

Image pull policy for Operator images.

Default: `IfNotPresent`

### `operator.imagePullSecrets`

List of the Image Pull Secrets for Operator images.

Default: `[]string`

### `operator.service.type`

Type of the Operator service.

Default: `ClusterIP`

### `operator.annotations`

Annotations passed to the Operator Deployment definition.

Default: `[]string`

### `operator.resources.limits.cpu`

CPU limits for operator pods.

Default: `1`

### `operator.resources.limits.memory`

Memory limits for operator pods.

Default: `256Mi`

### `operator.resources.requested.cpu`

Requested CPI by Operator pods.

Default: `250m`

### `operator.resources.requested.memory`

Requested memory for operator pods.

Default: `256Mi`

### `operator.replicaCount`

Replication count for Operator deployment.

Default: `2`

### `operator.updateStrategy`

Update strategy for operator pod.

Default: `Recreate`

### `operator.features.deployment`

Define if ArangoDeployment Operator should be enabled.

Default: `true`

### `operator.features.deploymentReplications`

Define if ArangoDeploymentReplications Operator should be enabled.

Default: `true`

### `operator.features.storage`

Define if ArangoLocalStorage Operator should be enabled.

Default: `false`

### `operator.features.backup`

Define if ArangoBackup Operator should be enabled.

Default: `false`

### `operator.enableCRDManagement`

If true and operator has enough access permissions, it will try to install missing CRDs.

Default: `true`

### `rbac.enabled`

Define if RBAC should be enabled.

Default: `true`

## Alternate namespaces

The `kube-arangodb` chart supports deployment into a non-default namespace.

To install the `kube-arangodb` chart is a non-default namespace, use the `--namespace`
argument like this.

```bash
helm install --namespace=mynamespace kube-arangodb.tgz
```

Note that since the operators claim exclusive access to a namespace, you can
install the `kube-arangodb` chart in a namespace once.
You can install the `kube-arangodb` chart in multiple namespaces. To do so, run:

```bash
helm install --namespace=namespace1 kube-arangodb.tgz
helm install --namespace=namespace2 kube-arangodb.tgz
```

The `kube-arangodb-storage` chart is always installed in the `kube-system` namespace.

## Common problems

### Error: no available release name found

This error is given by `helm install ...` in some cases where it has
insufficient permissions to install charts.

For various ways to work around this problem go to [this Stackoverflow article](https://stackoverflow.com/questions/43499971/helm-error-no-available-release-name-found).
