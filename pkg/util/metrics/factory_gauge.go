//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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

package metrics

import (
	"sync"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

type FactoryTypeGauge interface {
	comparable

	Desc() Description
	Gauge(value float64) Metric
}

func NewFactoryGauge[T FactoryTypeGauge]() FactoryGauge[T] {
	return &factoryGauge[T]{
		desc:    util.Default[T]().Desc(),
		metrics: map[T]float64{},
	}
}

type FactoryGauge[T FactoryTypeGauge] interface {
	Collector

	Items() []T
	Get(v T) float64
	Remove(v T)
	Set(v T, value float64)
	Add(v T, value float64)
}

type factoryGauge[T FactoryTypeGauge] struct {
	lock sync.Mutex

	desc    Description
	metrics map[T]float64
}

func (f *factoryGauge[T]) CollectMetrics(in PushMetric) {
	f.lock.Lock()
	defer f.lock.Unlock()

	for k, v := range f.metrics {
		in.Push(k.Gauge(v))
	}
}

func (f *factoryGauge[T]) CollectDescriptions(in PushDescription) {
	in.Push(f.desc)
}

func (f *factoryGauge[T]) Items() []T {
	f.lock.Lock()
	defer f.lock.Unlock()

	r := make([]T, 0, len(f.metrics))

	for k := range f.metrics {
		r = append(r, k)
	}

	return r
}

func (f *factoryGauge[T]) Get(v T) float64 {
	f.lock.Lock()
	defer f.lock.Unlock()

	return f.metrics[v]
}

func (f *factoryGauge[T]) Remove(v T) {
	f.lock.Lock()
	defer f.lock.Unlock()

	delete(f.metrics, v)
}

func (f *factoryGauge[T]) Set(v T, value float64) {
	f.lock.Lock()
	defer f.lock.Unlock()

	f.metrics[v] = value
}

func (f *factoryGauge[T]) Add(v T, value float64) {
	f.lock.Lock()
	defer f.lock.Unlock()

	f.metrics[v] = value + f.metrics[v]
}
