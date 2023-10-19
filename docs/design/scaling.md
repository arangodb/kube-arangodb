# Scaling

Number of running servers is controlled through `spec.<server_group>.count` field.

### Scale-up
When increasing the `count`, operator will try to create missing pods.
When scaling up make sure that you have enough computational resources / nodes, otherwise pod will stuck in Pending state.


### Scale-down

Scaling down is always done 1 server at a time.

Scale down is possible only when all other actions on ArangoDeployment are finished.

The internal process followed by the ArangoDB operator when scaling up is as follows:
- It chooses a member to be evicted. First it will try to remove unhealthy members or fall-back to the member with highest deletion_priority.  
- Making an internal calls, it forces the server to resign leadership.
  In case of DB servers it means that all shard leaders will be switched to other servers.
- Wait until server is cleaned out from cluster
- Pod finalized
