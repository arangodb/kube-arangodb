# Additional configuration

It is possible to additionally fine-tune operator behavior by
providing arguments via `operator.args` chart template value.

For example, you can specify burst size for k8s API requests or how long the operator
should wait for ArangoDeployment termination after receiving interruption signal:
```
operator:
  args: ["--kubernetes.burst=40", --shutdown.timeout=2m"]
```

The full list of available arguments can be retrieved using 
```
export OPERATOR_IMAGE=arangodb/kube-arangodb:1.2.9
kubectl run arango-operator-help --image=$OPERATOR_IMAGE -i --rm --restart=Never -- --help
```
