package metrics

import "github.com/prometheus/client_golang/prometheus"

type MetricApplier interface {
	ApplyMetrics(a MetricCollector)
}

type MetricCollector interface {
	Collect(m Metric, valueType prometheus.ValueType, value float64, labelValues ...string) MetricCollector

	CollectInstance(a MetricApplier) MetricCollector
}

type applier struct {
	metrics chan<- prometheus.Metric
}

func (a *applier) CollectInstance(app MetricApplier) MetricCollector {
	app.ApplyMetrics(a)
	return a
}

func (a *applier) Collect(m Metric, valueType prometheus.ValueType, value float64, labelValues ...string) MetricCollector {
	if metric, err := prometheus.NewConstMetric(m.Desc(), valueType, value, labelValues...); err == nil {
		a.metrics <- metric
	}

	return a
}

func NewMetricsCollector(metrics chan<- prometheus.Metric) MetricCollector {
	return &applier{metrics: metrics}
}
