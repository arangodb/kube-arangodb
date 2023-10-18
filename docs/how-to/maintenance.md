# Maintenance mode

## ArangoDeployment maintenance

When enabled, operator will pause reconciliation loop for specified ArangoDeployment.

Maintenance on ArangoDeployment can be enabled using annotation.

Key: `deployment.arangodb.com/maintenance`
Value: `true`

To enable maintenance mode for ArangoDeployment kubectl command can be used:
`kubectl annotate arangodeployment deployment deployment.arangodb.com/maintenance=true`

To disable maintenance mode for ArangoDeployment kubectl command can be used:
`kubectl annotate --overwrite arangodeployment deployment deployment.arangodb.com/maintenance-`

## Cluster maintenance

It is possible to put ArangoDB cluster into [agecy supervision mode](https://docs.arangodb.com/3.11/develop/http/cluster/#maintenance).

Use `spec.database.maintenance` field of ArangoDeployment CR to configure that:
```
spec:
  # ...
  database:
    maintenance: true

```
