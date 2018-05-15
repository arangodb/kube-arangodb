# Pod Eviction & Replacement

This chapter specifies the rules around evicting pods from nodes and
restarting or replacing them.

## Eviction

Eviction is the process of removing a pod that is running on a node from that node.

This is typically the result of a drain action (`kubectl drain`) or
from a taint being added to a node (either automatically by Kubernetes or manually by an operator).

## Replacement

Replacement is the process of replacing a pod by another pod that takes over the responsibilities
of the original pod.

The replacement pod has a new ID and new (read empty) persistent data.

Note that replacing a pod is different from restarting a pod. A pod is restarted when it has been reported
to have termined.

## NoExecute Tolerations

NoExecute tolerations are used to control the behavior of Kubernetes (wrt. to a Pod) when the node
that the pod is running on is no longer reachable or becomes not-ready.

See the applicable [Kubernetes documentation](https://kubernetes.io/docs/concepts/configuration/taint-and-toleration/) for more info.

## Rules

The rules for eviction & replacement are specified per type of pod.

### Image ID Pods

The Image ID pods are started to fetch the ArangoDB version of a specific
ArangoDB image and fetch the docker sha256 of that image.
They have no persistent state.

- Image ID pods can always be evicted from any node
- Image ID pods can always be restarted on a different node.
  There is no need to replace an image ID pod, nor will it cause problems when
  2 image ID pods run at the same time.
- `node.kubernetes.io/unreachable:NoExecute` toleration time is set very low (5sec)
- `node.kubernetes.io/not-ready:NoExecute` toleration time is set very low (5sec)

### Coordinator Pods

Coordinator pods run an ArangoDB coordinator as part of an ArangoDB cluster.
They have no persistent state, but do have a unique ID.

- Coordinator pods can always be evicted from any node
- Coordinator pods can always be replaced with another coordinator pod with a different ID on a different node
- `node.kubernetes.io/unreachable:NoExecute` toleration time is set low (15sec)
- `node.kubernetes.io/not-ready:NoExecute` toleration time is set low (15sec)

### DBServer Pods

DBServer pods run an ArangoDB dbserver as part of an ArangoDB cluster.
It has persistent state potentially tied to the node it runs on and it has a unique ID.

- DBServer pods can be evicted from any node as soon as:
  - It has been completely drained AND
  - It is no longer the shard master for any shard
- DBServer pods can be replaced with another dbserver pod with a different ID on a different node when:
  - It is not the shard master for any shard OR
  - For every shard it is the master for, there is an in-sync follower
- `node.kubernetes.io/unreachable:NoExecute` toleration time is set high to "wait it out a while" (5min)
- `node.kubernetes.io/not-ready:NoExecute` toleration time is set high to "wait it out a while" (5min)

### Agent Pods

Agent pods run an ArangoDB dbserver as part of an ArangoDB agency.
It has persistent state potentially tight to the node it runs on and it has a unique ID.

- Agent pods can be evicted from any node as soon as:
  - It is no longer the agency leader AND
  - There is at least an agency leader that is responding AND
  - There is at least an agency follower that is responding
- Agent pods can be replaced with another agent pod with the same ID but wiped persistent state on a different node when:
  - The old pod is known to be deleted (e.g. explicit eviction)
- `node.kubernetes.io/unreachable:NoExecute` toleration time is set high to "wait it out a while" (5min)
- `node.kubernetes.io/not-ready:NoExecute` toleration time is set high to "wait it out a while" (5min)

### Single Server Pods

Single server pods run an ArangoDB server as part of an ArangoDB single server deployment.
It has persistent state potentially tied to the node.

- Single server pods cannot be evicted from any node.
- Single server pods cannot be replaced with another pod.
- `node.kubernetes.io/unreachable:NoExecute` toleration time is not set to "wait it out forever"
- `node.kubernetes.io/not-ready:NoExecute` toleration time is not set "wait it out forever"

### Single Pods in Active Failover Deployment

Single pods run an ArangoDB single server as part of an ArangoDB active failover deployment.
It has persistent state potentially tied to the node it runs on and it has a unique ID.

- Single pods can be evicted from any node as soon as:
  - It is a follower of an active-failover deployment (Q: can we trigger this failover to another server?)
- Single pods can always be replaced with another single pod with a different ID on a different node.
- `node.kubernetes.io/unreachable:NoExecute` toleration time is set high to "wait it out a while" (5min)
- `node.kubernetes.io/not-ready:NoExecute` toleration time is set high to "wait it out a while" (5min)

### SyncMaster Pods

SyncMaster pods run an ArangoSync as master as part of an ArangoDB DC2DC cluster.
They have no persistent state, but do have a unique address.

- SyncMaster pods can always be evicted from any node
- SyncMaster pods can always be replaced with another syncmaster pod on a different node
- `node.kubernetes.io/unreachable:NoExecute` toleration time is set low (15sec)
- `node.kubernetes.io/not-ready:NoExecute` toleration time is set low (15sec)

### SyncWorker Pods

SyncWorker pods run an ArangoSync as worker as part of an ArangoDB DC2DC cluster.
They have no persistent state, but do have in-memory state and a unique address.

- SyncWorker pods can always be evicted from any node
- SyncWorker pods can always be replaced with another syncworker pod on a different node
- `node.kubernetes.io/unreachable:NoExecute` toleration time is set a bit higher to try to avoid resynchronization (1min)
- `node.kubernetes.io/not-ready:NoExecute` toleration time is set a bit higher to try to avoid resynchronization (1min)
