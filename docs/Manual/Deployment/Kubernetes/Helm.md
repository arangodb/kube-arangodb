# Using the ArangoDB Kubernetes Operator with Helm

[`Helm`](https://www.helm.sh/) is a package manager for Kubernetes, which enables
you to install various packages (include the ArangoDB Kubernetes Operator)
into your Kubernetes cluster.

The benefit of `helm` (in the context of the ArangoDB Kubernetes Operator)
is that it allows for a lot of flexibility in how you install the operator.
For example you can install the operator in a namespace other than
`default`.

## Charts

## Configurable values for ArangoDB Kubernetes Operator

The following values can be configured when installing the
ArangoDB Kubernetes Operator with `helm`.

### Both charts

| Key               | Type   | Description
|-------------------|--------|-----|
| Image             | string | Override the docker image used by the operators
| ImagePullPolicy   | string | Override the image pull policy used by the operators. See [Updating Images](https://kubernetes.io/docs/concepts/containers/images/#updating-images) for details.
| RBAC.Create       | bool   | Set to `true` (default) to create roles & role bindings.

### `kube-arangodb` chart

| Key               | Type   | Description
|-------------------|--------|-----|
| Deployment.Create | bool   | Set to `true` (default) to deploy the `ArangoDeployment` operator
| Deployment.User.ServiceAccountName | string | Name of the `ServiceAccount` that is the subject of the `RoleBinding` of users of the `ArangoDeployment` operator
| Deployment.Operator.ServiceAccountName | string | Name of the `ServiceAccount` used to run the `ArangoDeployment` operator
| Deployment.Operator.ServiceType | string | Type of `Service` created for the dashboard of the `ArangoDeployment` operator
| Deployment.AllowChaos | bool | Set to `true` to allow the introduction of chaos. **Only use for testing, never for production!** Defaults to `false`.
| DeploymentReplication.Create | bool   | Set to `true` (default) to deploy the `ArangoDeploymentReplication` operator
| DeploymentReplication.User.ServiceAccountName | string | Name of the `ServiceAccount` that is the subject of the `RoleBinding` of users of the `ArangoDeploymentReplication` operator
| DeploymentReplication.Operator.ServiceAccountName | string | Name of the `ServiceAccount` used to run the `ArangoDeploymentReplication` operator
| DeploymentReplication.Operator.ServiceType | string | Type of `Service` created for the dashboard of the `ArangoDeploymentReplication` operator

### `kube-arangodb-storage` chart

| Key               | Type   | Description
|-------------------|--------|-----|
| Storage.User.ServiceAccountName | string | Name of the `ServiceAccount` that is the subject of the `RoleBinding` of users of the `ArangoLocalStorage` operator
| Storage.Operator.ServiceAccountName | string | Name of the `ServiceAccount` used to run the `ArangoLocalStorage` operator
| Storage.Operator.ServiceType | string | Type of `Service` created for the dashboard of the `ArangoLocalStorage` operator
