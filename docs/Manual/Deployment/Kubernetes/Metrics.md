# Metrics

The ArangoDB Kubernetes Operator (`kube-arangodb`) exposes metrics of
its operations in a format that is compatible with [Prometheus](https://prometheus.io).

The metrics are exposed through HTTPS on port `8528` under path `/metrics`.

Look at [examples/metrics](https://github.com/arangodb/kube-arangodb/tree/master/examples/metrics)
for examples of `Services` and `ServiceMonitors` you can use to integrate
with Prometheus through the [Prometheus-Operator by CoreOS](https://github.com/coreos/prometheus-operator).

Furthermore, the operator can run sidecar containers for ArangoDB
deployments of type Cluster which expose metrics in Prometheus format.
Use the attribute `spec.metrics` to set this up see the [spec
reference](./DeploymentResource.md) for details.
