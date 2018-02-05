# Configuration & secrets

An ArangoDB cluster has lots of configuration options.
Some will be supported directly in the ArangoDB operator,
others will have to specified separately.

## Built-in options

All built-in options are passed to ArangoDB servers via commandline
arguments configured in the Pod-spec.

## Other configuration options

### Options 1

Use a `ConfigMap` per type of ArangoDB server.
The operator passes the options listed in the configmap
as commandline options to the ArangoDB servers.

TODO Discuss format of ConfigMap content. Is it `arangod.conf` like?

### Option 2

Add ArangoDB option sections to the custom resource.

## Secrets

The ArangoDB cluster needs several secrets such as JWT tokens
TLS certificates and so on.

All these secrets are stored as Kubernetes Secrets and passed to
the applicable Pods as files, mapped into the Pods filesystem.

The name of the secret is specified in the custom resource.
For example:

```yaml
apiVersion: "cluster.arangodb.com/v1alpha"
kind: "Cluster"
metadata:
  name: "example-arangodb-cluster"
spec:
  mode: cluster
  jwtTokenSecretName: <name-of-JWT-token-secret>
```
