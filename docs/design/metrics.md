# Metrics

Operator provides metrics in a format supported by [Prometheus](https://prometheus.io/).

The metrics are exposed via HTTP API. Default URL: `https://<operator-endpoint-ip>:8528/metrics`.

For a full list of available metrics, see [here](./../generated/metrics/README.md).


## Setting up integration with standard Prometheus installation (no TLS)

After creating operator deployment, you must configure Prometheus using a configuration file that instructs it
about which targets to scrape.
To do so, add a new scrape job to your prometheus.yaml config:
```yaml
scrape_configs:
  - job_name:       'arangodb-operator'

    scrape_interval: 10s # scrape every 10 seconds.

    scheme: 'https'
    # ignore TLS errors as operator uses self-signed certificate
    tls_config:
      insecure_skip_verify: true

    static_configs:
      - targets:
          - "<operator-endpoint-ip>:8528"
```

