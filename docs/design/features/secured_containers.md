# Secured Containers

## Overview

Change Default settings of:

* PodSecurityContext
  * `FSGroup` is set to `3000`
* SecurityContext (Container)
  * `RunAsUser` is set to `1000`
  * `RunAsGroup` is set to `2000`
  * `RunAsNonRoot` is set to `true`
  * `ReadOnlyRootFilesystem` is set to `true`
  * `Capabilities.Drop` is set to `["ALL"]`

## Dependencies

- [Operator Ephemeral Volumes](./ephemeral_volumes.md) should be Enabled and Supported. 

## How to use

To enable this feature use `--deployment.feature.secured-containers` arg, which needs be passed to the operator:

```shell
helm upgrade --install kube-arangodb \
https://github.com/arangodb/kube-arangodb/releases/download/$VER/kube-arangodb-$VER.tgz \
  --set "operator.args={--deployment.feature.secured-containers}"
```