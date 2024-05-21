//
// DISCLAIMER
//
// Copyright 2016-2024 ArangoDB GmbH, Cologne, Germany
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

package metric_descriptions

import (
	"sync"

	"github.com/arangodb/kube-arangodb/pkg/util/metrics"
)

var (
	arangoOperatorObjectsProcessed = metrics.NewDescription("arango_operator_objects_processed", "Number of the processed objects", []string{`operator_name`}, nil)
)

func init() {
	registerDescription(arangoOperatorObjectsProcessed)
	registerCollector(arangoOperatorObjectsProcessedGlobal)
}

func ArangoOperatorObjectsProcessed() metrics.Description {
	return arangoOperatorObjectsProcessed
}

func ArangoOperatorObjectsProcessedGet(operatorName string) float64 {
	return arangoOperatorObjectsProcessedGlobal.Get(ArangoOperatorObjectsProcessedItem{
		OperatorName: operatorName,
	})
}

func ArangoOperatorObjectsProcessedAdd(value float64, operatorName string) {
	arangoOperatorObjectsProcessedGlobal.Add(value, ArangoOperatorObjectsProcessedItem{
		OperatorName: operatorName,
	})
}

func ArangoOperatorObjectsProcessedInc(operatorName string) {
	arangoOperatorObjectsProcessedGlobal.Inc(ArangoOperatorObjectsProcessedItem{
		OperatorName: operatorName,
	})
}

func GetArangoOperatorObjectsProcessedFactory() ArangoOperatorObjectsProcessedFactory {
	return arangoOperatorObjectsProcessedGlobal
}

var arangoOperatorObjectsProcessedGlobal = &arangoOperatorObjectsProcessedFactory{
	items: arangoOperatorObjectsProcessedItems{},
}

type ArangoOperatorObjectsProcessedFactory interface {
	Get(object ArangoOperatorObjectsProcessedItem) float64
	Add(value float64, object ArangoOperatorObjectsProcessedItem)
	Remove(object ArangoOperatorObjectsProcessedItem)
	Items() []ArangoOperatorObjectsProcessedItem

	Inc(object ArangoOperatorObjectsProcessedItem)
}

type arangoOperatorObjectsProcessedFactory struct {
	lock sync.RWMutex

	items arangoOperatorObjectsProcessedItems
}

func (a *arangoOperatorObjectsProcessedFactory) Get(object ArangoOperatorObjectsProcessedItem) float64 {
	a.lock.Lock()
	defer a.lock.Unlock()

	v, ok := a.items[object]
	if !ok {
		return 0
	}

	return v
}

func (a *arangoOperatorObjectsProcessedFactory) Add(value float64, object ArangoOperatorObjectsProcessedItem) {
	a.lock.Lock()
	defer a.lock.Unlock()

	v, ok := a.items[object]
	if !ok {
		a.items[object] = value
		return
	}

	a.items[object] = value + v
}

func (a *arangoOperatorObjectsProcessedFactory) Remove(obj ArangoOperatorObjectsProcessedItem) {
	a.lock.Lock()
	defer a.lock.Unlock()

	delete(a.items, obj)
}

func (a *arangoOperatorObjectsProcessedFactory) Items() []ArangoOperatorObjectsProcessedItem {
	a.lock.Lock()
	defer a.lock.Unlock()

	var r = make([]ArangoOperatorObjectsProcessedItem, 0, len(a.items))

	for k := range a.items {
		r = append(r, k)
	}

	return r
}

func (a *arangoOperatorObjectsProcessedFactory) Inc(object ArangoOperatorObjectsProcessedItem) {
	a.Add(1, object)
}

func (a *arangoOperatorObjectsProcessedFactory) CollectMetrics(in metrics.PushMetric) {
	a.lock.RLock()
	defer a.lock.RUnlock()

	for k, v := range a.items {
		in.Push(arangoOperatorObjectsProcessed.Counter(v, k.OperatorName))
	}
}

func (a *arangoOperatorObjectsProcessedFactory) CollectDescriptions(in metrics.PushDescription) {
	in.Push(arangoOperatorObjectsProcessed)
}

type arangoOperatorObjectsProcessedItems map[ArangoOperatorObjectsProcessedItem]float64

type ArangoOperatorObjectsProcessedItem struct {
	OperatorName string
}
