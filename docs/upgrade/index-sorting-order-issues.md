---
layout: page
parent: Upgrading ArangoDB version
title: Index Sorting Order Issues
---

# Index Sorting Order Issues

ArangoDB References:
- [Resolving known issues with versions prior to 3.12.4](https://docs.arangodb.com/stable/release-notes/version-3.12/incompatible-changes-in-3-12/#resolving-known-issues-with-versions-prior-to-3124)
- [Corrected sorting order for numbers in VelocyPack indexes](https://docs.arangodb.com/stable/release-notes/version-3.12/incompatible-changes-in-3-12/#corrected-sorting-order-for-numbers-in-velocypack-indexes)

Feature: `--deployment.feature.upgrade-index-order-issue`

Affected Versions:
- 3.12.2
- 3.12.3

# Changes

During the upgrade Operator will change default [Member Upgrade Mode](../api/ArangoDeployment.V1.md#specagentsupgrademode) from `inplace` to `rotate` in order to recreate affected indexes.
