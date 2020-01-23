# Maintenance

## ArangoDeployment

Maintenance on ArangoDeployment can be enabled using annotation.

Key: `deployment.arangodb.com/maintenance`
Value: `true`

To enable maintenance mode for ArangoDeployment kubectl command can be used:
`kubectl annotate arangodeployment deployment deployment.arangodb.com/maintenance=true`

To disable maintenance mode for ArangoDeployment kubectl command can be used:
`kubectl annotate --overwrite arangodeployment deployment deployment.arangodb.com/maintenance-`