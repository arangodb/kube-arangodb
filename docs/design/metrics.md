# Metrics

## Integration with Prometheus Operator

Assuming that you have [Prometheus Operator](https://prometheus-operator.dev/) installed in your cluster (`monitoring` namespace),
and kube-arangodb installed in `default` namespace, you can easily configure the integration with ArangoDB operator.

The easiest way to do that is to create new a ServiceMonitor:
```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: arango-deployment-operator
  namespace: monitoring
  labels:
    prometheus: kube-prometheus
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: kube-arangodb
  namespaceSelector:
    matchNames:
    - default
  endpoints:
  - port: server
    scheme: https
    tlsConfig:
      insecureSkipVerify: true
```

You also can see the example of Grafana dashboard at `examples/metrics` folder of this repo.