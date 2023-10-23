# Operator Ephemeral Volumes

## Overview

Operator add 2 EmptyDir mounts to the ArangoDB Pods:

- `ephemeral-apps` which is mounted under `/ephemeral/app` and passed to the ArangoDB process via `--javascript.app-path` arg
- `ephemeral-tmp` which is mounted under `/ephemeral/tmp` and passed to the ArangoDB process via `--temp.path` arg

This adds possibility to enable ReadOnly FileSystem via PodSecurityContext configuration.

## How to use

To enable this feature use `--deployment.feature.ephemeral-volumes` arg, which needs be passed to the operator:

```shell
helm upgrade --install kube-arangodb \
https://github.com/arangodb/kube-arangodb/releases/download/$VER/kube-arangodb-$VER.tgz \
  --set "operator.args={--deployment.feature.ephemeral-volumes}"
```