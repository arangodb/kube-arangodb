# How to collect debug data

## Agency dump

To collect only agency dump, run:

```shell
kubectl exec -ti {POD_kube-arangodb-operator} -- /usr/bin/arangodb_operator admin agency dump > agency_dump.json
```

## Deployment debug package

To collect debug package, which contains things like:
- deployment pod logs
- operator pod logs
- kubernetes events
- deployment yaml files
- agency dump

Ensure you have debug mode enabled in the operator deployment:
```shell
```bash
helm upgrade --install kube-arangodb \
  https://github.com/arangodb/kube-arangodb/releases/download/$VER/kube-arangodb-$VER.tgz \
  --set "rbac.extensions.debug=true"
```
    
Then run:
```shell
kubectl exec -ti {POD_kube-arangodb-operator} -- /usr/bin/arangodb_operator debugPackage --namespace {namespace} -o - > db.tar.gz
```
