# Configuration & secrets

An ArangoDB cluster has lots of configuration options.
Some will be supported directly in the ArangoDB Operator,
others will have to specified separately.

## Passing command line options

All command-line options of `arangod` (and `arangosync`) are available
by adding options to the `spec.<group>.args` list of a group
of servers.

These arguments are added to the command-line created for these servers.

## Secrets

The ArangoDB cluster needs several secrets such as JWT tokens
TLS certificates and so on.

All these secrets are stored as Kubernetes Secrets and passed to
the applicable Pods as files, mapped into the Pods filesystem.

The name of the secret is specified in the custom resource.
For example:

```yaml
apiVersion: "database.arangodb.com/v1"
kind: "ArangoDeployment"
metadata:
  name: "example-simple-cluster"
spec:
  mode: Cluster
  image: 'arangodb/arangodb:3.10.8'
  auth:
    jwtSecretName: <name-of-JWT-token-secret>
```
