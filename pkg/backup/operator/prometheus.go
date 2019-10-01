package operator

import (
	"github.com/prometheus/client_golang/prometheus"
)

type prometheusMetrics struct {
	operator *operator

	objectProcessed prometheus.Counter
}

func newCollector(operator *operator) *prometheusMetrics {
	return &prometheusMetrics{
		operator: operator,

		objectProcessed: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "arango_operator_objects_processed",
			Help: "Count of the processed objects",
			ConstLabels: map[string]string{
				"operator_name": operator.name,
			},
		}),
	}
}

func (p *prometheusMetrics) connectors() []prometheus.Collector {
	return []prometheus.Collector{
		p.objectProcessed,
	}
}

func (p *prometheusMetrics) Describe(r chan<- *prometheus.Desc) {
	for _, c := range p.connectors() {
		c.Describe(r)
	}

	for _, h := range p.operator.handlers {
		if collector, ok := h.(prometheus.Collector); ok {
			collector.Describe(r)
		}
	}
}

func (p *prometheusMetrics) Collect(r chan<- prometheus.Metric) {
	for _, c := range p.connectors() {
		c.Collect(r)
	}

	for _, h := range p.operator.handlers {
		if collector, ok := h.(prometheus.Collector); ok {
			collector.Collect(r)
		}
	}
}
