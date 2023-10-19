# ArangoDeployment Custom Resource

[Full CustomResourceDefinition reference ->](./api/ArangoDeployment.V1.md)

The ArangoDB Deployment Operator creates and maintains ArangoDB deployments
in a Kubernetes cluster, given a deployment specification.
This deployment specification is a `CustomResource` following
a `CustomResourceDefinition` created by the operator.

Example minimal deployment definition of an ArangoDB database cluster:

```yaml
apiVersion: "database.arangodb.com/v1"
kind: "ArangoDeployment"
metadata:
  name: "example-arangodb-cluster"
spec:
  mode: Cluster
```

Example more elaborate deployment definition:

```yaml
apiVersion: "database.arangodb.com/v1"
kind: "ArangoDeployment"
metadata:
  name: "example-arangodb-cluster"
spec:
  mode: Cluster
  environment: Production
  agents:
    count: 3
    args:
      - --log.level=debug
    resources:
      requests:
        storage: 8Gi
    storageClassName: ssd
  dbservers:
    count: 5
    resources:
      requests:
        storage: 80Gi
    storageClassName: ssd
  coordinators:
    count: 3
  image: "arangodb/arangodb:3.9.3"
```
