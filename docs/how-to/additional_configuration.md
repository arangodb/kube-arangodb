# How to pass additional params to operator

It is possible to additionally fine-tune operator behavior by
providing arguments via `operator.args` chart template value.

The full list of available arguments can be retrieved using
```
export OPERATOR_IMAGE=arangodb/kube-arangodb:$VER
kubectl run arango-operator-help --image=$OPERATOR_IMAGE -i --rm --restart=Never -- --help
```


### Example 1: kubernetes.burst 

You can specify burst size for k8s API requests or how long the operator
should wait for ArangoDeployment termination after receiving interruption signal:
```
operator:
  args: ["--kubernetes.burst=40", --shutdown.timeout=2m"]
```

### Example 2: CRD validation

You can specify which of installed CRD should have a validation schema enabled:
```
operator:
  args:
    - --crd.validation-schema=arangobackuppolicies.backup.arangodb.com=true
    - --crd.validation-schema=arangodeployments.database.arangodb.com=false
```
