# Restore defaults from last accepted state of deployment

## Overview

ArangoDeployment has a lot of fields, which have default values.
If `--deployment.feature.deployment-spec-defaults-restore` is enabled (which is by default),
then the operator will restore the default values from the last accepted state of the deployment.

E.g., if user removes the `spec.dbservers` field from the deployment,
then the operator will restore the default value of this field back.

## How to use

To disable this feature use `--deployment.feature.deployment-spec-defaults-restore=false` arg, which needs be passed to the operator:

```shell
helm upgrade --install kube-arangodb \
https://github.com/arangodb/kube-arangodb/releases/download/$VER/kube-arangodb-$VER.tgz \
  --set "operator.args={--deployment.feature.deployment-spec-defaults-restore=false}"
```
