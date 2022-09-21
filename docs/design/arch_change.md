# Member Architecture change

To change manually architecture of specific member, you can use annotation:
```bash
kubectl annotate pod arango-pod deployment.arangodb.com/arch=arm64
```

It will add to the member status `ArchitectureMismatch` condition, e.g.:
```yaml
  - lastTransitionTime: "2022-09-15T07:38:05Z"
    lastUpdateTime: "2022-09-15T07:38:05Z"
    reason: Member has a different architecture than the deployment
    status: "True"
    type: ArchitectureMismatch
```

To provide requested arch changes for the member we need rotate it, so additional step is required:
```bash
`kubectl annotate pod arango-pod deployment.arangodb.com/rotate=true`
```
