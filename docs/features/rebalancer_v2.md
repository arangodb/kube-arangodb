# ArangoDB Rebalancer V2 Support

## Overview

ArangoDB as of 3.10.0 provides Cluster Rebalancer functionality via [api](https://www.arangodb.com/docs/stable/http/cluster.html#rebalance).

Operator will use the above functionality to check shard movement plan and enforce it on the Cluster.


## How to use

To enable this feature use `--deployment.feature.rebalancer-v2` arg, which needs be passed to the operator:

```shell
helm upgrade --install kube-arangodb \
https://github.com/arangodb/kube-arangodb/releases/download/$VER/kube-arangodb-$VER.tgz \
  --set "operator.args={--deployment.feature.rebalancer-v2}"
```

To enable Rebalancer in ArangoDeployment:
```yaml
spec:
   rebalancer:
     enabled: true
```