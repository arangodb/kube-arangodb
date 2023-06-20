# Failover Leader service

## Overview

This feature is designed to solve the problem with the Leader service in Active Failover mode.
It attaches the `deployment.arangodb.com/leader=true` label to the Leader member of the cluster.
If a member is a Follower, then this label is removed.
Above labels are used by the cluster Services to route the traffic to the Leader member.

In case of double Leader situation (which will be fixed in future versions of ArangoDB), 
the operator will remove the `deployment.arangodb.com/leader=true` label from all members, 
which will cause the cluster outage.

## How to use

This feature is disabled by default.
To enable it use `--deployment.feature.failover-leadership ` arg, which needs be passed to the operator:

```shell
helm upgrade --install kube-arangodb \
https://github.com/arangodb/kube-arangodb/releases/download/$VER/kube-arangodb-$VER.tgz \
  --set "operator.args={--deployment.feature.force-rebuild-out-synced-shards}"
```
