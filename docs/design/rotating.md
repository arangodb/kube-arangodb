# Rotation

## ArangoDeployment

Rotation of ArangoDeployment Pods can be triggered by Pod deletion or by annotation (safe way).

Using annotation Pods gonna be rotated one-by-one which will keep cluster alive.

Key: `deployment.arangodb.com/rotate`
Value: `true`

To rotate ArangoDeployment Pod kubectl command can be used:
`kubectl annotate pod arango-pod deployment.arangodb.com/rotate=true`
