package metrics

import "github.com/prometheus/client_golang/prometheus"

type Counter interface {
	Add()

	MetricApplier
	MetricDesc
}

type counter struct {
	Metric
	labels []string

	count int64
}

func (c *counter) Add() {
	c.count++
}

func (c *counter) ApplyMetrics(a MetricCollector) {
	a.Collect(c, prometheus.CounterValue, float64(c.count), c.labels...)
}

type MetricDesc interface {
	Desc() *prometheus.Desc
}

type Metric interface {
	MetricDesc

	NewCounter(labels ...string) Counter
}

func NewMetricDescription(fqName, help string, variableLabels []string) Metric {
	return metricDescription{prometheus.NewDesc(fqName, help, variableLabels, nil)}
}

type metricDescription struct {
	Description *prometheus.Desc
}

func (m metricDescription) NewCounter(labels ...string) Counter {
	return &counter{
		Metric: m,
		count:  0,
		labels: labels,
	}
}

func (m metricDescription) Desc() *prometheus.Desc {
	return m.Description
}
