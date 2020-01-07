# Acceptance test for kube-arangodb operator on specific Kubernetes platform

This acceptance test plan describes all test scenario's that must be executed
successfully in order to consider the kube-arangodb operator production ready
on a specific Kubernetes setup (from now on we'll call a Kubernetes setup a platform).

## Platform parameters

Before the test, record the following parameters for the platform the test is executed on.

- Name of the platform
- Version of the platform
- Upstream Kubernetes version used by the platform (run `kubectl version`)
- Number of nodes used by the Kubernetes cluster (run `kubectl get node`)
- `StorageClasses` provided by the platform (run `kubectl get storageclass`)
- Does the platform use RBAC? (run `kubectl describe clusterrolebinding`)
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

To do so, follow the [instructions in the documentation](https://www.arangodb.com/docs/stable/deployment-kubernetes-usage.html).

### `PersistentVolume` provider

If the platform does not provide a `PersistentVolume` provider, create one by running:

```bash
kubectl apply -f examples/arango-local-storage.yaml
```

## Basis tests

The basis tests are executed on every platform with various images:

Run the following tests with the following images:

- Community <Version>
- Enterprise <Version>

For every tests, one of these images can be chosen, as long as each image
is used in a test at least once.

### Test 1a: Create single server deployment

Create an `ArangoDeployment` of mode `Single`.

Hint: Use `tests/acceptance/single.yaml`.

- [ ] The deployment must start
- [ ] The deployment must yield 1 `Pod`
- [ ] The deployment must yield a `Service` named `<deployment-name>`
- [ ] The deployment must yield a `Service` named `<deployment-name>-ea`
- [ ] The `Service` named `<deployment-name>-ea` must be accessible from outside (LoadBalancer or NodePort) and show WebUI

### Test 1b: Create active failover deployment

Create an `ArangoDeployment` of mode `ActiveFailover`.

Hint: Use `tests/acceptance/activefailover.yaml`.

- [ ] The deployment must start
- [ ] The deployment must yield 5 `Pods`
- [ ] The deployment must yield a `Service` named `<deployment-name>`
- [ ] The deployment must yield a `Service` named `<deployment-name>-ea`
- [ ] The `Service` named `<deployment-name>-ea` must be accessible from outside (LoadBalancer or NodePort) and show WebUI

### Test 1c: Create cluster deployment

Create an `ArangoDeployment` of mode `Cluster`.

Hint: Use `tests/acceptance/cluster.yaml`.

- [ ] The deployment must start
- [ ] The deployment must yield 9 `Pods`
- [ ] The deployment must yield a `Service` named `<deployment-name>`
- [ ] The deployment must yield a `Service` named `<deployment-name>-ea`
- [ ] The `Service` named `<deployment-name>-ea` must be accessible from outside (LoadBalancer or NodePort) and show WebUI

### Test 1d: Create cluster deployment with dc2dc

This test requires the use of the enterprise image.

Create an `ArangoDeployment` of mode `Cluster` and dc2dc enabled.

Hint: Derive from `tests/acceptance/cluster-sync.yaml`.

- [ ] The deployment must start
- [ ] The deployment must yield 15 `Pods`
- [ ] The deployment must yield a `Service` named `<deployment-name>`
- [ ] The deployment must yield a `Service` named `<deployment-name>-ea`
- [ ] The deployment must yield a `Service` named `<deployment-name>-sync`
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

Hint: Use `tests/acceptance/cluster.yaml`.

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

Hint: Derive from `tests/acceptance/single.yaml`.

- [ ] The deployment must start
- [ ] The deployment must yield 1 `Pod`
- [ ] The deployment must yield a `Service` named `<deployment-name>`
- [ ] The deployment must yield a `Service` named `<deployment-name>-ea`
- [ ] The `Service` named `<deployment-name>-ea` must be accessible from outside (LoadBalancer or NodePort) and show WebUI

### Test 3b: Create active failover deployment in production environment

Create an `ArangoDeployment` of mode `ActiveFailover` with an environment of `Production`.

Hint: Derive from `tests/acceptance/activefailover.yaml`.

- [ ] The deployment must start
- [ ] The deployment must yield 5 `Pods`
- [ ] The deployment must yield a `Service` named `<deployment-name>`
- [ ] The deployment must yield a `Service` named `<deployment-name>-ea`
- [ ] The `Service` named `<deployment-name>-ea` must be accessible from outside (LoadBalancer or NodePort) and show WebUI

### Test 3c: Create cluster deployment in production environment

Create an `ArangoDeployment` of mode `Cluster` with an environment of `Production`.

Hint: Derive from `tests/acceptance/cluster.yaml`.

- [ ] The deployment must start
- [ ] The deployment must yield 9 `Pods`
- [ ] The deployment must yield a `Service` named `<deployment-name>`
- [ ] The deployment must yield a `Service` named `<deployment-name>-ea`
- [ ] The `Service` named `<deployment-name>-ea` must be accessible from outside (LoadBalancer or NodePort) and show WebUI

### Test 3d: Create cluster deployment in production environment and scale it

Create an `ArangoDeployment` of mode `Cluster` with an environment of `Production`.

Hint: Derive from `tests/acceptance/cluster.yaml`.

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

Change the value of `spec.coordinators.count` from 4 to 2.

- [ ] Three coordinators are removed (one by one)
- [ ] The deployment must yield 7 `Pods`

### Test 4a: Create cluster deployment with `ArangoLocalStorage` provided volumes

Ensure an `ArangoLocalStorage` is deployed.

Hint: Use from `tests/acceptance/local-storage.yaml`.

Create an `ArangoDeployment` of mode `Cluster` with a `StorageClass` that is
mapped to an `ArangoLocalStorage` provider.

Hint: Derive from `tests/acceptance/cluster.yaml`.

- [ ] The deployment must start
- [ ] The deployment must yield 9 `Pods`
- [ ] The deployment must yield a `Service` named `<deployment-name>`
- [ ] The deployment must yield a `Service` named `<deployment-name>-ea`
- [ ] The `Service` named `<deployment-name>-ea` must be accessible from outside (LoadBalancer or NodePort) and show WebUI

### Test 4b: Create cluster deployment with a platform provided `StorageClass`

This test only applies to platforms that provide their own `StorageClasses`.

Create an `ArangoDeployment` of mode `Cluster` with a `StorageClass` that is
provided by the platform.

Hint: Derive from `tests/acceptance/cluster.yaml`.

- [ ] The deployment must start
- [ ] The deployment must yield 9 `Pods`
- [ ] The deployment must yield a `Service` named `<deployment-name>`
- [ ] The deployment must yield a `Service` named `<deployment-name>-ea`
- [ ] The `Service` named `<deployment-name>-ea` must be accessible from outside (LoadBalancer or NodePort) and show WebUI

### Test 5a: Test `Pod` resilience on single servers

Create an `ArangoDeployment` of mode `Single`.

Hint: Use from `tests/acceptance/single.yaml`.

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

Hint: Use from `tests/acceptance/activefailover.yaml`.

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

Hint: Use from `tests/acceptance/cluster.yaml`.

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

### Test 6a: Test `Node` reboot on single servers

Create an `ArangoDeployment` of mode `Single`.

Hint: Use from `tests/acceptance/single.yaml`.

- [ ] The deployment must start
- [ ] The deployment must yield 1 `Pod`
- [ ] The deployment must yield a `Service` named `<deployment-name>`
- [ ] The deployment must yield a `Service` named `<deployment-name>-ea`
- [ ] The `Service` named `<deployment-name>-ea` must be accessible from outside (LoadBalancer or NodePort) and show WebUI

Reboot the `Node` of the deployment that contains the single server.

- [ ] The `Pod` running on the `Node` must be restarted
- [ ] After the `Pod` has restarted, the server must have the same data and be responsive again

### Test 6b: Test `Node` reboot on active failover

Create an `ArangoDeployment` of mode `ActiveFailover` with an environment of `Production`.

Hint: Use from `tests/acceptance/activefailover.yaml`.

- [ ] The deployment must start
- [ ] The deployment must yield 5 `Pods`
- [ ] The deployment must yield a `Service` named `<deployment-name>`
- [ ] The deployment must yield a `Service` named `<deployment-name>-ea`
- [ ] The `Service` named `<deployment-name>-ea` must be accessible from outside (LoadBalancer or NodePort) and show WebUI

Reboot a `Node`.

- [ ] While the `Node` is restarting, the cluster must still respond to requests (R/W)
- [ ] All `Pods` on the `Node` must be restarted

### Test 6c: Test `Node` reboot on clusters

Create an `ArangoDeployment` of mode `Cluster` with an environment of `Production`.

Hint: Use from `tests/acceptance/cluster.yaml`.

- [ ] The deployment must start
- [ ] The deployment must yield 9 `Pods`
- [ ] The deployment must yield a `Service` named `<deployment-name>`
- [ ] The deployment must yield a `Service` named `<deployment-name>-ea`
- [ ] The `Service` named `<deployment-name>-ea` must be accessible from outside (LoadBalancer or NodePort) and show WebUI

Reboot a `Node`.

- [ ] While the `Node` is restarting, the cluster must still respond to requests (R/W)
- [ ] All `Pods` on the `Node` must be restarted

### Test 6d: Test `Node` removal on single servers

This test is only valid when `StorageClass` is used that provides network attached `PersistentVolumes`.

Create an `ArangoDeployment` of mode `Single`.

Hint: Use from `tests/acceptance/single.yaml`.

- [ ] The deployment must start
- [ ] The deployment must yield 1 `Pod`
- [ ] The deployment must yield a `Service` named `<deployment-name>`
- [ ] The deployment must yield a `Service` named `<deployment-name>-ea`
- [ ] The `Service` named `<deployment-name>-ea` must be accessible from outside (LoadBalancer or NodePort) and show WebUI

Remove the `Node` containing the deployment from the Kubernetes cluster.

- [ ] The `Pod` running on the `Node` must be restarted on another `Node`
- [ ] After the `Pod` has restarted, the server must have the same data and be responsive again

### Test 6e: Test `Node` removal on active failover

Create an `ArangoDeployment` of mode `ActiveFailover` with an environment of `Production`.

Hint: Use from `tests/acceptance/activefailover.yaml`.

- [ ] The deployment must start
- [ ] The deployment must yield 5 `Pods`
- [ ] The deployment must yield a `Service` named `<deployment-name>`
- [ ] The deployment must yield a `Service` named `<deployment-name>-ea`
- [ ] The `Service` named `<deployment-name>-ea` must be accessible from outside (LoadBalancer or NodePort) and show WebUI

Remove a `Node` containing the `Pods` of the deployment from the Kubernetes cluster.

- [ ] While the `Pods` are being restarted on new `Nodes`, the cluster must still respond to requests (R/W)
- [ ] The `Pods` running on the `Node` must be restarted on another `Node`
- [ ] After the `Pods` have restarted, the server must have the same data and be responsive again

### Test 6f: Test `Node` removal on clusters

This test is only valid when:

- A `StorageClass` is used that provides network attached `PersistentVolumes`
- or all collections have a replication factor of 2 or higher

Create an `ArangoDeployment` of mode `Cluster` with an environment of `Production`.

Hint: Use from `tests/acceptance/cluster.yaml`.

- [ ] The deployment must start
- [ ] The deployment must yield 9 `Pods`
- [ ] The deployment must yield a `Service` named `<deployment-name>`
- [ ] The deployment must yield a `Service` named `<deployment-name>-ea`
- [ ] The `Service` named `<deployment-name>-ea` must be accessible from outside (LoadBalancer or NodePort) and show WebUI

Remove a `Node` containing the `Pods` of the deployment from the Kubernetes cluster.

- [ ] While the `Pods` are being restarted on new `Nodes`, the cluster must still respond to requests (R/W)
- [ ] The `Pods` running on the `Node` must be restarted on another `Node`
- [ ] After the `Pods` have restarted, the server must have the same data and be responsive again

### Test 6g: Test `Node` removal on clusters with replication factor 1

This test is only valid when:

- A `StorageClass` is used that provides `Node` local `PersistentVolumes`
- and at least some collections have a replication factor of 1

Create an `ArangoDeployment` of mode `Cluster` with an environment of `Production`.

Hint: Use from `tests/acceptance/cluster.yaml`.

- [ ] The deployment must start
- [ ] The deployment must yield 9 `Pods`
- [ ] The deployment must yield a `Service` named `<deployment-name>`
- [ ] The deployment must yield a `Service` named `<deployment-name>-ea`
- [ ] The `Service` named `<deployment-name>-ea` must be accessible from outside (LoadBalancer or NodePort) and show WebUI

Remove a `Node`, containing the dbserver `Pod` that holds a collection with replication factor 1,
from the Kubernetes cluster.

- [ ] While the `Pods` are being restarted on new `Nodes`, the cluster must still respond to requests (R/W),
      except requests involving collections with a replication factor of 1
- [ ] The `Pod` running the dbserver with a collection that has a replication factor of 1 must NOT be restarted on another `Node`

Remove the collections with the replication factor of 1

- [ ] The remaining `Pods` running on the `Node` must be restarted on another `Node`
- [ ] After the `Pods` have restarted, the server must have the same data, except for the removed collections, and be responsive again

### Test 7a: Test DC2DC on 2 clusters, running in the same Kubernetes cluster

This test requires the use of the enterprise image.

Create 2 `ArangoDeployment` of mode `Cluster` and dc2dc enabled.

Hint: Derive from `tests/acceptance/cluster-sync.yaml`, name the deployments `cluster1` and `cluster2`.

Make sure to include a name ('cluster1-to-2`) for an external access package.

```yaml
apiVersion: "database.arangodb.com/v1alpha"
kind: "ArangoDeployment"
metadata:
  name: "cluster1"
spec:
  mode: Cluster
  image: ewoutp/arangodb:3.3.14
  sync:
    enabled: true
    externalAccess:
      accessPackageSecretNames: ["cluster1-to-2"]
```

- [ ] The deployments must start
- [ ] The deployments must yield 15 `Pods`
- [ ] The deployments must yield a `Service` named `cluster[1|2]`
- [ ] The deployments must yield a `Service` named `cluster[1|2]-ea`
- [ ] The deployments must yield a `Service` named `cluster[1|2]-sync`
- [ ] The `Services` named `cluster[1|2]-ea` must be accessible from outside (LoadBalancer or NodePort) and show WebUI

Create an `ArangoDeploymentReplication` from `tests/acceptance/cluster12-replication.yaml`.

It will take some time until the synchronization (from `cluster1` to `cluster2`) is configured.

- [ ] The status of the `cluster12-replication` resource shows ....
- [ ] The webUI of `cluster1` shows that you can create a new collection there.
- [ ] The webUI of `cluster2` shows that you cannot create a new collection there.

Create a collection named `testcol` with a replication factor 2 and 3 shards (using the webUI of `cluster1`).

- [ ] The webUI of `cluster2` shows collection `testcol` with the given replication factor and number of shards.

Create multiple documents in the collection named `testcol` (using the webUI of `cluster1`).

- [ ] The documents are visible in webUI of `cluster2`.

Modify multiple documents in the collection named `testcol` (using the webUI of `cluster1`).

- [ ] The modified documents are visible in webUI of `cluster2`.

Remove one or more documents from the collection named `testcol` (using the webUI of `cluster1`).

- [ ] The documents are no longer visible in webUI of `cluster2`.

Create a new database called `db2` (using the webUI of `cluster1`).

- [ ] The webUI of `cluster2` shows database `db2`.
