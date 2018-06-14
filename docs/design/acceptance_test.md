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

### Test 3: Production environment

Production environment tests are only relevant if there are enough nodes
available that `Pods` can be scheduled on.

The number of available nodes must be >= the maximum server count in
any group.

### Test 3a: Create single server deployment in production environment

Create an `ArangoDeployment` of mode `Single` with an environment of `Production`.

- [ ] The deployment must start
- [ ] The deployment must yield 1 `Pod`
- [ ] The deployment must yield a `Service` named `<deployment-name>`
- [ ] The deployment must yield a `Service` named `<deployment-name>-ea`
- [ ] The `Service` named `<deployment-name>-ea` must be accessible from outside (LoadBalancer or NodePort) and show WebUI

### Test 3b: Create active failover deployment in production environment

Create an `ArangoDeployment` of mode `ActiveFailover` with an environment of `Production`.

- [ ] The deployment must start
- [ ] The deployment must yield 5 `Pods`
- [ ] The deployment must yield a `Service` named `<deployment-name>`
- [ ] The deployment must yield a `Service` named `<deployment-name>-ea`
- [ ] The `Service` named `<deployment-name>-ea` must be accessible from outside (LoadBalancer or NodePort) and show WebUI

### Test 3c: Create cluster deployment in production environment

Create an `ArangoDeployment` of mode `Cluster` with an environment of `Production`.

- [ ] The deployment must start
- [ ] The deployment must yield 9 `Pods`
- [ ] The deployment must yield a `Service` named `<deployment-name>`
- [ ] The deployment must yield a `Service` named `<deployment-name>-ea`
- [ ] The `Service` named `<deployment-name>-ea` must be accessible from outside (LoadBalancer or NodePort) and show WebUI

### Test 3d: Create cluster deployment in production environment and scale it

Create an `ArangoDeployment` of mode `Cluster` with an environment of `Production`.

- [ ] The deployment must start
- [ ] The deployment must yield 9 `Pods`
- [ ] The deployment must yield a `Service` named `<deployment-name>`
- [ ] The deployment must yield a `Service` named `<deployment-name>-ea`
- [ ] The `Service` named `<deployment-name>-ea` must be accessible from outside (LoadBalancer or NodePort) and show WebUI

Change the value of `spec.dbservers.count` from 3 to 4.

- [ ] Two dbservers are added
- [ ] The deployment must yield 10 `Pods`

Change the value of `spec.coordinators.count` from 3 to 4.

- [ ] A coordinator is added
- [ ] The deployment must yield 11 `Pods`

Change the value of `spec.dbservers.count` from 4 to 2.

- [ ] Three dbservers are removed (one by one)
- [ ] The deployment must yield 9 `Pods`

Change the value of `spec.coordinators.count` from 4 to 1.

- [ ] Three coordinators are removed (one by one)
- [ ] The deployment must yield 6 `Pods`

### Test 4a: Create cluster deployment with `ArangoLocalStorage` provided volumes

Ensure an `ArangoLocalStorage` is deployed.

Create an `ArangoDeployment` of mode `Cluster` with a `StorageClass` that is
mapped to an `ArangoLocalStorage` provider.

- [ ] The deployment must start
- [ ] The deployment must yield 9 `Pods`
- [ ] The deployment must yield a `Service` named `<deployment-name>`
- [ ] The deployment must yield a `Service` named `<deployment-name>-ea`
- [ ] The `Service` named `<deployment-name>-ea` must be accessible from outside (LoadBalancer or NodePort) and show WebUI

### Test 4b: Create cluster deployment with a platform provides `StorageClass`

This test only applies to platforms that provide their own `StorageClasses`.

Create an `ArangoDeployment` of mode `Cluster` with a `StorageClass` that is
provided by the platform.

- [ ] The deployment must start
- [ ] The deployment must yield 9 `Pods`
- [ ] The deployment must yield a `Service` named `<deployment-name>`
- [ ] The deployment must yield a `Service` named `<deployment-name>-ea`
- [ ] The `Service` named `<deployment-name>-ea` must be accessible from outside (LoadBalancer or NodePort) and show WebUI

### Test 5a: Test `Pod` resilience on single servers

Create an `ArangoDeployment` of mode `Single`.

- [ ] The deployment must start
- [ ] The deployment must yield 1 `Pod`
- [ ] The deployment must yield a `Service` named `<deployment-name>`
- [ ] The deployment must yield a `Service` named `<deployment-name>-ea`
- [ ] The `Service` named `<deployment-name>-ea` must be accessible from outside (LoadBalancer or NodePort) and show WebUI

Delete the `Pod` of the deployment that contains the single server.

- [ ] The `Pod` must be restarted
- [ ] After the `Pod` has restarted, the server must have the same data and be responsive again

### Test 5b: Test `Pod` resilience on active failover

Create an `ArangoDeployment` of mode `ActiveFailover`.

- [ ] The deployment must start
- [ ] The deployment must yield 5 `Pods`
- [ ] The deployment must yield a `Service` named `<deployment-name>`
- [ ] The deployment must yield a `Service` named `<deployment-name>-ea`
- [ ] The `Service` named `<deployment-name>-ea` must be accessible from outside (LoadBalancer or NodePort) and show WebUI

Delete a `Pod` of the deployment that contains an agent.

- [ ] While the `Pod` is gone & restarted, the cluster must still respond to requests (R/W)
- [ ] The `Pod` must be restarted

Delete a `Pod` of the deployment that contains a single server.

- [ ] While the `Pod` is gone & restarted, the cluster must still respond to requests (R/W)
- [ ] The `Pod` must be restarted

### Test 5c: Test `Pod` resilience on clusters

Create an `ArangoDeployment` of mode `Cluster`.

- [ ] The deployment must start
- [ ] The deployment must yield 9 `Pods`
- [ ] The deployment must yield a `Service` named `<deployment-name>`
- [ ] The deployment must yield a `Service` named `<deployment-name>-ea`
- [ ] The `Service` named `<deployment-name>-ea` must be accessible from outside (LoadBalancer or NodePort) and show WebUI

Delete a `Pod` of the deployment that contains an agent.

- [ ] While the `Pod` is gone & restarted, the cluster must still respond to requests (R/W)
- [ ] The `Pod` must be restarted

Delete a `Pod` of the deployment that contains a dbserver.

- [ ] While the `Pod` is gone & restarted, the cluster must still respond to requests (R/W), except
      for requests to collections with a replication factor of 1.
- [ ] The `Pod` must be restarted

Delete a `Pod` of the deployment that contains an coordinator.

- [ ] While the `Pod` is gone & restarted, the cluster must still respond to requests (R/W), except
      requests targeting the restarting coordinator.
- [ ] The `Pod` must be restarted

## Further ideas to be discussed

I just collect further things which I think are missing:

  - add resilience tests:
     - reboot a node, should come back, at least if nothing is ephemeral
     - kill a node permanently with replicated data, should recover and repair
     - kill a node if it contains non-replicated data
       should hang and not recover, but dropping the collection should
       alow it to recover and repair (obviously, without the data)
