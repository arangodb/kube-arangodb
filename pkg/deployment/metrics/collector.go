package metrics

type Collector interface {
	Collect(metrics MetricCollector) error
}
