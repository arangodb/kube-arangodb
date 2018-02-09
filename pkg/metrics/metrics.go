//
// DISCLAIMER
//
// Copyright 2018 ArangoDB GmbH, Cologne, Germany
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Copyright holder is ArangoDB GmbH, Cologne, Germany
//
// Author Ewout Prangsma
//

package metrics

import "github.com/prometheus/client_golang/prometheus"

const (
	namespace = "arangodb_operator"
)

// MustRegisterCounter creates and registers a counter.
// Must be called from `init`.
func MustRegisterCounter(component, name, help string) prometheus.Counter {
	m := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: component,
		Name:      name,
		Help:      help,
	})
	prometheus.MustRegister(m)
	return m
}

// MustRegisterCounterVec creates and registers a counter vector.
// Must be called from `init`.
func MustRegisterCounterVec(component, name, help string, labelNames ...string) *prometheus.CounterVec {
	m := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: component,
		Name:      name,
		Help:      help,
	}, labelNames)
	prometheus.MustRegister(m)
	return m
}

// MustRegisterGauge creates and registers a gauge.
// Must be called from `init`.
func MustRegisterGauge(component, name, help string) prometheus.Gauge {
	m := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: component,
		Name:      name,
		Help:      help,
	})
	prometheus.MustRegister(m)
	return m
}

// MustRegisterGaugeVec creates and registers a gauge vector.
// Must be called from `init`.
func MustRegisterGaugeVec(component, name, help string, labelNames ...string) *prometheus.GaugeVec {
	m := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: component,
		Name:      name,
		Help:      help,
	}, labelNames)
	prometheus.MustRegister(m)
	return m
}

// MustRegisterSummary creates and registers a summary.
// Must be called from `init`.
func MustRegisterSummary(component, name, help string, objectives map[float64]float64) prometheus.Summary {
	if objectives == nil {
		objectives = map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001}
	}
	m := prometheus.NewSummary(prometheus.SummaryOpts{
		Namespace:  namespace,
		Subsystem:  component,
		Name:       name,
		Help:       help,
		Objectives: objectives,
	})
	prometheus.MustRegister(m)
	return m
}
