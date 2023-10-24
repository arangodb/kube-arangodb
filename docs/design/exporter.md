# ArangoDB Exporter for Prometheus

This exporter exposes the statistics provided by a specific ArangoDB instance
in a format compatible with prometheus.

## Usage

To use the ArangoDB Exporter, run the following:

```bash
arangodb_operator exporter \
    --arangodb.endpoint=http://<your-database-host>:8529 \
    --arangodb.jwtsecret=<your-jwt-secret> \
    --ssl.keyfile=<your-optional-ssl-keyfile>
```

This results in an ArangoDB Exporter exposing all statistics of
the ArangoDB server (running at `http://<your-database-host>:8529`)
at `http://<your-host-ip>:9101/metrics`.

## Exporter mode

Expose ArangoDB metrics for ArangoDB >= 3.6.0

In default mode metrics provided by ArangoDB `_admin/metrics` (<=3.7) or `_admin/metrics/v2` (3.8+) are exposed on Exporter port.

## Configuring Prometheus

There are several ways to configure Prometheus to fetch metrics from the ArangoDB Exporter.

Below you will find a sample Prometheus configuration file that can be used to fetch
metrics from an ArangoDB exporter listening on localhost port 9101 (without TLS).

```yaml
global:
  scrape_interval:     15s
scrape_configs:
- job_name: arangodb
  static_configs:
  - targets: ['localhost:9101']
```

For more info on configuring Prometheus go to [its configuration documentation](https://prometheus.io/docs/prometheus/latest/configuration/configuration).

If you're using the [Prometheus Operator](https://github.com/coreos/prometheus-operator)
in Kubernetes, you need to create an additional `Service` and a `ServiceMonitor` resource
like this:

```yaml
kind: Service
apiVersion: v1
metadata:
  name: arangodb-exporters-service
  labels:
    app: arangodb-exporter
spec:
  selector:
    app: arangodb-exporter
  ports:
  - name: metrics
    port: 9101

---

apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: arangodb-exporter
  namespace: monitoring
  labels:
    team: frontend
    prometheus: kube-prometheus
spec:
  namespaceSelector:
    matchNames:
    - default
  selector:
    matchLabels:
      app: arangodb-exporter
  endpoints:
  - port: metrics
    scheme: https
    tlsConfig:
      insecureSkipVerify: true
```

Note 1: that the typical deployment on the Prometheus operator is done in
a namespace called `monitoring`. Make sure to match the `namespace`
of the `ServiceMonitor` to match that.

Note 2: that the `Prometheus` custom resource has a field called `serviceMonitorSelector`.
Make sure that the `matchLabels` selector in there matches the labels of
your `ServiceMonitor`.