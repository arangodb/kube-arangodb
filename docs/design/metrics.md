# Metrics

Operator provides metrics of its operations in a format supported by [Prometheus](https://prometheus.io/).

The metrics are exposed through HTTPS on port `8528` under path `/metrics`.

For a full list of available metrics, see [here](./../generated/metrics/README.md).

#### Contents
- [Integration with standard Prometheus installation (no TLS)](#Integration-with-standard-Prometheus-installation-no-TLS)
- [Integration with Prometheus Operator](#Integration-with-Prometheus-Operator)
- [Exposing ArangoDB metrics](#ArangoDB-metrics)


## Integration with standard Prometheus installation (no TLS)

After creating operator deployment, you must configure Prometheus using a configuration file that instructs it
about which targets to scrape.
To do so, add a new scrape job to your prometheus.yaml config:
```yaml
scrape_configs:
  - job_name:       'arangodb-operator'

    scrape_interval: 10s # scrape every 10 seconds.

    scheme: 'https'
    tls_config:
      insecure_skip_verify: true

    static_configs:
      - targets:
          - "<operator-endpoint-ip>:8528"
```

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



## ArangoDB metrics

The operator can run sidecar containers for ArangoDB deployments of type `Cluster` which expose metrics in Prometheus format.
Edit your `ArangoDeployment` resource, setting `spec.metrics.enabled` to true to enable ArangoDB metrics.
The operator will run a sidecar container for every cluster component.
In addition to the sidecar containers the operator will deploy a `Service` to access the exporter ports (from within the k8s cluster),
and a resource of type `ServiceMonitor`, provided the corresponding custom resource definition is deployed in the k8s cluster.
If you are running Prometheus in the same k8s cluster with the Prometheus operator, this will be the case.
The ServiceMonitor will have the following labels set:
```yaml
app: arangodb
arango_deployment: YOUR_DEPLOYMENT_NAME
context: metrics
metrics: prometheus
```
This makes it possible to configure your Prometheus deployment to automatically start monitoring on the available Prometheus feeds.
To this end, you must configure the `serviceMonitorSelector` in the specs of your Prometheus deployment to match these labels. For example:
```yaml
serviceMonitorSelector:
    matchLabels:
      metrics: prometheus
```
would automatically select all pods of all ArangoDB cluster deployments which have metrics enabled.

See the [list of exposed ArangoDB metrics](https://www.arangodb.com/docs/stable/http/administration-and-monitoring-metrics.html#list-of-exposed-metrics)
