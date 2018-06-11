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

### Test 1: Create single server deployment

Create an `ArangoDeployment` of mode `Single`.

- [ ] The deployment must start
- [ ] The deployment 

## Scenario's

The following test scenario's must be covered by automated tests:

- Creating 1 deployment (all modes, all environments, all storage engines)
- Creating multiple deployments (all modes, all environments, all storage engines),
  controlling each individually
- Creating deployment with/without authentication
- Creating deployment with/without TLS

- Updating deployment wrt:
  - Number of servers (scaling, up/down)
  - Image version (upgrading, downgrading within same minor version range (e.g. 3.2.x))
  - Immutable fields (should be reset automatically)

- Resilience:
  - Delete individual pods
  - Delete individual PVCs
  - Delete individual Services
  - Delete Node
  - Restart Node
  - API server unavailable

- Persistent Volumes:
  - hint: RBAC file might need to be changed
  - hint: get info via - client-go.CoreV1()
  - Number of volumes should stay in reasonable bounds
  - For some cases it might be possible to check that, the amount before and after the test stays the same
  - A Cluster start should need 6 Volumes (DBServer + Agents)
  - The release of a volume-claim should result in a release of the volume

## Test environments

- Kubernetes clusters
  - Single node
  - Multi node
  - Access control mode (RBAC, ...)
  - Persistent volumes ...
