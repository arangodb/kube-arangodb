package metrics

import (
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Collector interface {
	Collect(chan<- prometheus.Collector)
}

type CollectorFactory interface {
	Collector
	CollectFactory(chan<- Collector)
}

func WrapHttpForCollectors(collectors ...Collector) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		r := prometheus.NewRegistry()
		Register(r, collectors...)

		promhttp.HandlerFor(r, promhttp.HandlerOpts{}).ServeHTTP(writer, request)
	}
}

func Register(registerer prometheus.Registerer, collectors ...Collector) {
	c := make(chan prometheus.Collector)

	go func() {
		defer close(c)
		Collect(c, collectors...)
	}()

	for m := range c {
		registerer.Register(m)
	}
}

func Collect(c chan<- prometheus.Collector, collectors ...Collector) {
	var wg sync.WaitGroup

	for _, collector := range collectors {
		wg.Add(1)
		go func() {
			defer wg.Done()
			collectSingle(c, collector)
		}()
	}

	wg.Wait()
}

func collectSingle(c chan<- prometheus.Collector, collector Collector) {
	collector.Collect(c)

	if f, ok := collector.(CollectorFactory); ok {
		collectFactory(c, f)
	}

}

func collectFactory(c chan<- prometheus.Collector, f CollectorFactory) {
	collectors := make(chan Collector)

	go func() {
		defer close(collectors)
		f.CollectFactory(collectors)
	}()

	var wg sync.WaitGroup

	for m := range collectors {
		wg.Add(1)

		go func() {
			defer wg.Done()

			collectSingle(c, m)
		}()
	}

	wg.Wait()
}
