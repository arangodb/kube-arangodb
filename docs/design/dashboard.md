# Deployment Operator Dashboard

To inspect the state of an `ArangoDeployment` you can use `kubectl get ...` to inspect
the `status` of the resource itself, but to get the entire "picture" you also
must inspect the status of the `Pods` created for the deployment, the `PersistentVolumeClaims`,
the `PersistentVolumes`, the `Services` and some `Secrets`.

The goal of the operator dashboard is to simplify this inspection process.

The deployment operator dashboard provides:

- A status overview of all `ArangoDeployments` it controls
- A status overview of all resources created by the operator (for an `ArangoDeployment`)
- Run the arangoinspector on deployments
- Instructions for upgrading deployments to newer versions

It does not provide:

- Direct access to the deployed database
- Anything that can already be done in the web-UI of the database or naturaly belongs there.

The dashboard is a single-page web application that is served by the operator itself.

## Design decisions

### Leader only

Since only the operator instance that won the leader election has the latest state of all
deployments, only that instance will serve dashboard requests.

For this purpose, a `Service` is created when deploying the operator.
This service uses a `role=leader` selector to ensure that only the right instance
will be included in its list of endpoints.

### Exposing the dashboard

By default the `Service` that selects the leading operator instance is not exposed outside the Kubernetes cluster.
Users must use `kubectl expose service ...` to add additional `Services` of type `LoadBalancer`
or `NodePort` to expose the dashboard if and how they want to.

### Readonly behavior

The dashboard only provides readonly functions.
When modifications to an `ArangoDeployment` are needed (e.g. when upgrading to a new version), the dashboard
will provide instructions for doing so using `kubectl` commands.

In doing so, the requirements for authentication & access control of the dashboard itself remain limited,
while all possible authentication & access control features of Kubernetes are still available to ensure
a secure deployment.

### Authentication

The dashboard requires a username+password to gain access, unless it is started with an option to disable authentication.
This username+password pair is stored in a standard basic authentication `Secret` in the Kubernetes cluster.

### Frontend technology

The frontend part of the dashboard will be built with React.
This aligns with future developments in the context of the web-UI of the database itself.

### Backend technology

The backend of the dashboard contains an HTTPS server that serves the dashboard webpage (including all required web resources)
and all API methods it needs.
