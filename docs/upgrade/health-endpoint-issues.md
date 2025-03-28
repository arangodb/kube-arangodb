---
layout: page
parent: Upgrading ArangoDB version
title: Coordinator Health Endpoint Issue
---

# Coordinator Health Endpoint Issue

Affected Versions:
- < 3.12.4

# Changes

During the upgrade Operator will change default [Upgrade Order](../api/ArangoDeployment.V1.md#specupgradeorder) from `standard` to `coordinatorFirst` in order to update coordinators first.
