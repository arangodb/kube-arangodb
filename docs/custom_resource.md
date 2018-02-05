# Custom Resource

The ArangoDB operator creates and maintains ArangoDB clusters
in a Kubernetes cluster, given a cluster specification.
This cluster specification is a CustomResource following
a CustomResourceDefinition created by the operator.

Example minimal cluster definition:

```yaml
apiVersion: "cluster.arangodb.com/v1alpha"
kind: "Cluster"
metadata:
  name: "example-arangodb-cluster"
spec:
  mode: cluster
```

Example more elaborate cluster definition:

```yaml
apiVersion: "cluster.arangodb.com/v1alpha"
kind: "Cluster"
metadata:
  name: "example-arangodb-cluster"
spec:
  mode: cluster
  agents:
    servers: 3
    args:
      - --log.level=debug
    resources:
      requests:
        storage: 8Gi
    storageClassName: ssd
  dbservers:
    servers: 5
    resources:
      requests:
        storage: 80Gi
    storageClassName: ssd
  coordinators:
    servers: 3
  image: "arangodb/arangodb:3.3.3"
```
