---
layout: page
title: Helm
parent: ArangoDBPlatform
nav_order: 4
---

# Helm Details

## Labels

### installation.platform.arangodb.com/managed

Always set to true - defines if Helm Release is managed by the ArangoDB Platform tools

### installation.platform.arangodb.com/chart

Defines the name of the chart used for the installation

### installation.platform.arangodb.com/deployment

Defines the ArangoDeployment name

### installation.platform.arangodb.com/service

Optional. Defines the service used to spawn Helm Release

### installation.platform.arangodb.com/type

Defines type if installation. Possible values:

- platform - managed by the ArangoDBPlatform Installer
- service - managed by the services
