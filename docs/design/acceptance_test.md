# Acceptance test for kube-arangodb operator on specific Kubernetes platform

This acceptance test plan describes all test scenario's that must be executed
succesfully in order to consider the kube-arangodb operator production ready
on a specific Kubernetes setup (from now on we'll call a Kubernetes setup a platform).

## Platform parameters

Before the test, record the following parameters for the platform the test is executed on.

- Name of the platform
- Version of the platform
- Upstream Kubernetes version used by the platform
- Number of nodes used by the Kubernetes cluster
- `StorageClasses` provided by the platform (run `kubectl get storageclass`)
- Does the platform use RBAC?
- Does the platform support services of type `LoadBalancer`?

If one of the above questions can have multiple answers (e.g. different Kubernetes versions)
then make the platform more specific. E.g. consider "GKE with Kubernetes 1.10.2" a platform
instead of "GKE" which can have version "1.8", "1.9" & "1.10.2".

## Platform preparations

Before the tests can be run, the platform has to be prepared.

### Deploy the ArangoDB operators

Deploy the following ArangoDB operators:

- `ArangoDeployment` operator
- `ArangoDeploymentReplication` operator
- `ArangoLocalStorage` operator

To do so, follow the [instructions in the manual](../Manual/Deployment/Kubernetes/Usage.md).

### `PersistentVolume` provider

If the platform does not provide a `PersistentVolume` provider, create one by running:

```bash
kubectl apply -f examples/arango-local-storage.yaml
```

## Basis tests

The basis tests are executed on every platform with various images:

Run the following tests for the following images:

- Community 3.3.10
- Enterprise 3.3.10

### Test 1a: Create single server deployment

Create an `ArangoDeployment` of mode `Single`.

- [ ] The deployment must start
- [ ] The deployment must yield 1 `Pod`
- [ ] The deployment must yield a `Service` named `<deployment-name>`
- [ ] The deployment must yield a `Service` named `<deployment-name>-ea`
- [ ] The `Service` named `<deployment-name>-ea` must be accessible from outside (LoadBalancer or NodePort) and show WebUI

### Test 1b: Create active failover deployment

Create an `ArangoDeployment` of mode `ActiveFailover`.

- [ ] The deployment must start
- [ ] The deployment must yield 5 `Pods`
- [ ] The deployment must yield a `Service` named `<deployment-name>`
- [ ] The deployment must yield a `Service` named `<deployment-name>-ea`
- [ ] The `Service` named `<deployment-name>-ea` must be accessible from outside (LoadBalancer or NodePort) and show WebUI

### Test 1c: Create cluster deployment

Create an `ArangoDeployment` of mode `Cluster`.

- [ ] The deployment must start
- [ ] The deployment must yield 9 `Pods`
- [ ] The deployment must yield a `Service` named `<deployment-name>`
- [ ] The deployment must yield a `Service` named `<deployment-name>-ea`
- [ ] The `Service` named `<deployment-name>-ea` must be accessible from outside (LoadBalancer or NodePort) and show WebUI

### Test 2a: Scale an active failover deployment

Create an `ArangoDeployment` of mode `ActiveFailover`.

- [ ] The deployment must start
- [ ] The deployment must yield 5 `Pods`
- [ ] The deployment must yield a `Service` named `<deployment-name>`
- [ ] The deployment must yield a `Service` named `<deployment-name>-ea`
- [ ] The `Service` named `<deployment-name>-ea` must be accessible from outside (LoadBalancer or NodePort) and show WebUI

Change the value of `spec.single.count` from 2 to 3.

- [ ] A single server is added
- [ ] The deployment must yield 6 `Pods`

Change the value of `spec.single.count` from 3 to 2.

- [ ] A single server is removed
- [ ] The deployment must yield 5 `Pods`

### Test 2b: Scale a cluster deployment

Create an `ArangoDeployment` of mode `Cluster`.

- [ ] The deployment must start
- [ ] The deployment must yield 9 `Pods`
- [ ] The deployment must yield a `Service` named `<deployment-name>`
- [ ] The deployment must yield a `Service` named `<deployment-name>-ea`
- [ ] The `Service` named `<deployment-name>-ea` must be accessible from outside (LoadBalancer or NodePort) and show WebUI

Change the value of `spec.dbservers.count` from 3 to 5.

- [ ] Two dbservers are added
- [ ] The deployment must yield 11 `Pods`

Change the value of `spec.coordinators.count` from 3 to 4.

- [ ] A coordinator is added
- [ ] The deployment must yield 12 `Pods`

Change the value of `spec.dbservers.count` from 5 to 2.

- [ ] Three dbservers are removed (one by one)
- [ ] The deployment must yield 9 `Pods`

Change the value of `spec.coordinators.count` from 4 to 1.

- [ ] Three coordinators are removed (one by one)
- [ ] The deployment must yield 6 `Pods`
