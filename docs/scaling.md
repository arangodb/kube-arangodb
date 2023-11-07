# Scaling your ArangoDB deployment

The ArangoDB Kubernetes Operator allows easily scale the number of DB-Servers and Coordinators up or down as needed.

The scale up or down, change the number of servers in the custom resource.

E.g., change `spec.dbservers.count` from `3` to `4`.

Then apply the updated resource using:

```bash
kubectl apply -f {your-arango-deployment}.yaml
```

Inspect the status of the custom resource to monitor the progress of the scaling operation.

**Note: It is not possible to change the number of Agency servers after creating a cluster**.
Make sure to specify the desired number when creating CR first time.


## Overview

### Scale-up

When increasing the `count`, operator will try to create missing pods.
When scaling up, make sure that you have enough computational resources / nodes, otherwise pod will be stuck in Pending state.


### Scale-down

Scaling down is always done one server at a time.

Scale down is possible only when all other actions on ArangoDeployment are finished.

The internal process followed by the ArangoDB operator when scaling down is as follows:
- It chooses a member to be evicted. First, it will try to remove unhealthy members or fall-back to the member with 
  the highest `deletion_priority` (check [Use deletion_priority to control scale-down order](#use-deletion_priority-to-control-scale-down-order)).
- Making an internal calls, it forces the server to resign leadership.
  In case of DB servers it means that all shard leaders will be switched to other servers.
- Wait until server is cleaned out from the cluster.
- Pod finalized.

#### Use deletion_priority to control scale-down order

You can use `.spec.deletion_priority` field in `ArangoMember` CR to control the order in which servers are scaled down.
Refer to [ArangoMember API Reference](/docs/api/ArangoMember.V1.md#specdeletionpriority-integer) for more details.
