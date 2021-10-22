# Introduction

Initial RBAC for namespace-wide access used by arangodb operator deployment

# Chart Details

Chart will install fully namespace admin ServiceAccount which will allow to deploy arangodb operator in particular namespace.

# Prerequisites

Cluster admin role is required to deploy this chart.

# Installing the Chart

Run following command using helm3
```
helm install arango-init . --namespace demo-jwierzbo --create-namespace --wait
```

To update existing Helm deployment:
```helm upgrade arango-init . --namespace demo-jwierzbo --wait```
