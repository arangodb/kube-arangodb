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
// Author Adam Janikowski
//

package operator

import (
	"fmt"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type Starter interface {
	Start(stopCh <-chan struct{})
}

type Operator interface {
	// Implement prometheus collector interface
	prometheus.Collector

	Name() string

	Start(threadiness int, stopCh <-chan struct{}) error

	RegisterInformer(informer cache.SharedIndexInformer, group, version, kind string) error
	RegisterStarter(starter Starter) error
	RegisterHandler(handler Handler) error

	EnqueueItem(item Item)
	ProcessItem(item Item) error
}

func NewOperator(name string) Operator {
	o := &operator{
		name:      name,
		workqueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), name),
	}

	o.prometheusMetrics = newCollector(o)

	return o
}

type operator struct {
	lock sync.Mutex

	started bool

	name string

	informers []cache.SharedInformer
	starters  []Starter
	handlers  []Handler

	workqueue workqueue.RateLimitingInterface

	// Implement prometheus collector
	*prometheusMetrics
}

func (o *operator) Name() string {
	return o.name
}

func (o *operator) ProcessItem(item Item) error {
	{
		o.lock.Lock()
		defer o.lock.Unlock()

		if !o.started {
			return fmt.Errorf("operator is not started started")
		}
	}

	return o.processItem(item)
}

func (o *operator) RegisterHandler(handler Handler) error {
	o.lock.Lock()
	defer o.lock.Unlock()

	if o.started {
		return fmt.Errorf("operator already started")
	}

	for _, registeredHandlers := range o.handlers {
		if registeredHandlers == handler {
			return fmt.Errorf("handler already registered")
		}
	}

	o.handlers = append(o.handlers, handler)

	return nil
}

func (o *operator) RegisterStarter(starter Starter) error {
	o.lock.Lock()
	defer o.lock.Unlock()

	if o.started {
		return fmt.Errorf("operator already started")
	}

	for _, registeredStarter := range o.starters {
		if registeredStarter == starter {
			return fmt.Errorf("starter already registered")
		}
	}

	o.starters = append(o.starters, starter)

	return nil
}

func (o *operator) EnqueueItem(item Item) {
	o.workqueue.Add(item.String())
}

func (o *operator) RegisterInformer(informer cache.SharedIndexInformer, group, version, kind string) error {
	o.lock.Lock()
	defer o.lock.Unlock()

	if o.started {
		return fmt.Errorf("operator already started")
	}

	for _, registeredInformer := range o.informers {
		if registeredInformer == informer {
			return fmt.Errorf("informer already registered")
		}
	}

	o.informers = append(o.informers, informer)

	informer.AddEventHandler(newResourceEventHandler(o, group, version, kind))

	return nil
}

func (o *operator) Start(threadiness int, stopCh <-chan struct{}) error {
	o.lock.Lock()
	defer o.lock.Unlock()

	if o.started {
		return fmt.Errorf("operator already started")
	}

	o.started = true

	return o.start(threadiness, stopCh)
}

func (o *operator) start(threadiness int, stopCh <-chan struct{}) error {
	// Execute pre checks
	log.Info().Msgf("Executing Lifecycle PreStart")
	for _, handler := range o.handlers {
		if err := ExecLifecyclePreStart(handler); err != nil {
			return err
		}
	}

	log.Info().Msgf("Starting informers")
	for _, starter := range o.starters {
		starter.Start(stopCh)
	}

	if err := o.waitForCacheSync(stopCh); err != nil {
		return err
	}

	log.Info().Msgf("Starting workers")
	for i := 0; i < threadiness; i++ {
		go wait.Until(o.worker, time.Second, stopCh)
	}

	log.Info().Msgf("Operator started")
	return nil
}

func (o *operator) waitForCacheSync(stopCh <-chan struct{}) error {
	cacheSync := make([]cache.InformerSynced, len(o.informers))

	for id, informer := range o.informers {
		cacheSync[id] = informer.HasSynced
	}

	if ok := cache.WaitForCacheSync(stopCh, cacheSync...); !ok {
		return fmt.Errorf("cache can not sync")
	}

	return nil
}
