# Metrics

TODO:

- Investigate prometheus annotations wrt metrics
    - see https://github.com/prometheus/prometheus/blob/master/documentation/examples/prometheus-kubernetes.yml
    - `prometheus.io/scrape`: Only scrape services that have a value of `true`
    - `prometheus.io/scheme`: If the metrics endpoint is secured then you will need
    - `prometheus.io/path`: If the metrics path is not `/metrics` override this.
    - `prometheus.io/port`: If the metrics are exposed on a different port to the

- Add prometheus compatible `/metrics` endpoint to `arangod`
