package metrics

import "github.com/prometheus/client_golang/prometheus"

type DescCollector interface {
	Collect(m MetricDesc) DescCollector
}

func NewDesc(descs chan<- *prometheus.Desc) DescCollector {
	return &desc{
		descs: descs,
	}
}

type desc struct {
	descs chan<- *prometheus.Desc
}

func (d *desc) Collect(m MetricDesc) DescCollector {
	d.descs <- m.Desc()
	return d
}
