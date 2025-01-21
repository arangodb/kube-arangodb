---
layout: page
title: How to replace Pod
parent: How to ...
---

# How to replace Pod

Replacement of ArangoDeployment Pods can be triggered by annotation.

Replacement of the pod will ensure that destruction of the member will never result in data loss.

Replacement is disabled for:
- Single Servers

Replacement of member for particular groups will result in:
- Agents - Member and PVC is recreated, Operator ensures that Quorum is keept during this operation
- DBServers - New DBServer is added where data is migrated before removal
- Coordinator, Gateway - Simple shutdown

Key: `deployment.arangodb.com/replace`
Value: `true`

To rotate ArangoDeployment Pod kubectl command can be used:
`kubectl annotate pod arango-pod deployment.arangodb.com/replace=true`
