# Architecture change

Currently `AMD64` is a default architecture in the operator
To enable `ARM64` support in operator add following config in kube-arangodb chart:

```bash
helm upgrade --install kube-arangodb \
  https://github.com/arangodb/kube-arangodb/releases/download/$VER/kube-arangodb-$VER.tgz \
  --set "operator.architectures={amd64,arm64}"
```

## ARM64 ArangoDeployment

`AMD64` is a default architecture in the ArangoDeployment.
`ARM64` is available since ArangoDB 3.10.0.
To create `ARM64` ArangoDeployment you need to add `arm64` architecture to the deployment:

```yaml
apiVersion: database.arangodb.com/v1
kind: ArangoDeployment
metadata:
  name: cluster
spec:
  image: arangodb:3.10
  mode: Cluster
  architecture:
    - arm64
```

## Member Architecture change (Enterprise only)

To migrate members from AMD64 to ARM64 you need to add `arm64` architecture to the existing deployment as a first item on the list, e.g.
```yaml
apiVersion: database.arangodb.com/v1
kind: ArangoDeployment
metadata:
  name: cluster
spec:
  image: arangodb:3.10
  mode: Cluster
  architecture:
    - arm64
    - amd64
```

All new members since now will be created on ARM64 nodes
All recreated members since now will be created on ARM64 nodes

To change architecture of a specific member, you need to use following annotation:
```bash
kubectl annotate pod {MY_POD} deployment.arangodb.com/arch=arm64
```

It will add to the member status `ArchitectureMismatch` condition, e.g.:
```yaml
  - lastTransitionTime: "2022-09-15T07:38:05Z"
    lastUpdateTime: "2022-09-15T07:38:05Z"
    reason: Member has a different architecture than the deployment
    status: "True"
    type: ArchitectureMismatch
```

To provide requested arch changes for the member, we need to rotate it, so an additional step is required:
```bash
`kubectl annotate pod {MY_POD} deployment.arangodb.com/rotate=true`
```
