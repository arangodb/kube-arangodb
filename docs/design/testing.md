# Testing

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
