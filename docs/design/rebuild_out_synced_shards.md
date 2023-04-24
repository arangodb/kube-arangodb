# Force rebuild out-synced Shards with broken Merkle Tree

## Overview

TODO

## How to use

This feature is disabled by default. To enable it `--deployment.feature.force-rebuild-out-synced-shards` arg needs be passed to the operator.
We can also change default timeout value (60 min) for this feature by `--timeout.shard-rebuild {duration}` arg.

Here is the example `helm` command which enables this feature and sets timeout to 10 minutes:
```shell
export VER=1.2.27; helm upgrade --install kube-arangodb \
https://github.com/arangodb/kube-arangodb/releases/download/$VER/kube-arangodb-$VER.tgz \
  --set "operator.args={--deployment.feature.force-rebuild-out-synced-shards,--timeout.shard-rebuild=10m}"
```
