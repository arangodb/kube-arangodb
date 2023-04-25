# Force rebuild out-synced Shards with broken Merkle Tree

## Overview

TODO

## How to use

This feature is disabled by default. 
- To enable it use `--deployment.feature.force-rebuild-out-synced-shards` arg, which needs be passed to the operator.
- Optionally we can override default timeouts by attaching following args to the operator:
  - `--timeout.shard-rebuild {duration}` - timeout after which particular out-synced shard is considered as failed and rebuild is triggered (default 60m0s)
  - `--timeout.shard-rebuild-retry {duration}` - timeout after which rebuild shards retry flow is triggered (default 60m0s)

Here is the example `helm` command which enables this feature and sets shard-rebuild timeout to 10 minutes:
```shell
export VER=1.2.27; helm upgrade --install kube-arangodb \
https://github.com/arangodb/kube-arangodb/releases/download/$VER/kube-arangodb-$VER.tgz \
  --set "operator.args={--deployment.feature.force-rebuild-out-synced-shards,--timeout.shard-rebuild=10m}"
```
