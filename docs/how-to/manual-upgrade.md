---
layout: page
title: How to run upgrade manually
parent: How to ...
---

# How to upgrade deployment manually (within Maintenance windows)

## Upgrade Mode Change

In order to upgrade Members manually, Upgrade Mode needs to be changed on the ArangoDeployment Level to `manual`.

Use `spec.dbservers.upgradeMode` field of ArangoDeployment CR to configure that:
```
spec:
  # ...
  dbservers:
    upgradeMode: manual

```

## Trigger upgrade

In order to trigger upgrade annotation `upgrade.deployment.arangodb.com/allow` needs to be set on the Pod.

Kubectl command: `kubectl annotate pod arango-pod upgrade.deployment.arangodb.com/allow=true`