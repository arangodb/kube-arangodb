# How to rotate Pod

Rotation of ArangoDeployment Pods can be triggered by Pod deletion or by annotation (safe way).

Using annotation is preferred way to rotate pods while keeping cluster in health state.

Key: `deployment.arangodb.com/rotate`
Value: `true`

To rotate ArangoDeployment Pod kubectl command can be used:
`kubectl annotate pod arango-pod deployment.arangodb.com/rotate=true`
