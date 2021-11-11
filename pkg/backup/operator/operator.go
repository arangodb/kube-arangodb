//
// DISCLAIMER
//
// Copyright 2016-2021 ArangoDB GmbH, Cologne, Germany
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
	"sync"
	"time"

	"github.com/rs/zerolog"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"

	"github.com/arangodb/kube-arangodb/pkg/backup/operator/operation"

	"github.com/prometheus/client_golang/prometheus"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

// Starter interface used by Operator to start new GoRoutines
type Starter interface {
	Start(stopCh <-chan struct{})
}

// Operator interface for operator core functionality
type Operator interface {
	// Define prometheus collector interface
	prometheus.Collector

	Name() string
	Namespace() string
	Image() string

	Start(threadiness int, stopCh <-chan struct{}) error

	RegisterInformer(informer cache.SharedIndexInformer, group, version, kind string) error
	RegisterStarter(starter Starter) error
	RegisterHandler(handler Handler) error

	EnqueueItem(item operation.Item)
	ProcessItem(item operation.Item) error

	GetLogger() *zerolog.Logger
}

// NewOperator creates new operator
func NewOperator(logger zerolog.Logger, name, namespace, image string) Operator {
	o := &operator{
		name:      name,
		namespace: namespace,
		image:     image,
		logger:    logger,
		workqueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), name),
	}

	// Declaration of prometheus interface
	o.prometheusMetrics = newCollector(o)

	return o
}

type operator struct {
	lock sync.Mutex

	started bool

	logger zerolog.Logger

	name      string
	namespace string
	image     string

	informers []cache.SharedInformer
	starters  []Starter
	handlers  []Handler

	workqueue workqueue.RateLimitingInterface

	// Implement prometheus collector
	*prometheusMetrics
}

func (o *operator) Namespace() string {
	return o.namespace
}

func (o *operator) Name() string {
	return o.name
}

func (o *operator) Image() string {
	return o.image
}

func (o *operator) ProcessItem(item operation.Item) error {
	{
		o.lock.Lock()
		defer o.lock.Unlock()

		if !o.started {
			return errors.Newf("operator is not started started")
		}
	}

	return o.processItem(item)
}

func (o *operator) RegisterHandler(handler Handler) error {
	o.lock.Lock()
	defer o.lock.Unlock()

	if o.started {
		return errors.Newf("operator already started")
	}

	for _, registeredHandlers := range o.handlers {
		if registeredHandlers == handler {
			return errors.Newf("handler already registered")
		}
	}

	o.handlers = append(o.handlers, handler)

	return nil
}

func (o *operator) RegisterStarter(starter Starter) error {
	o.lock.Lock()
	defer o.lock.Unlock()

	if o.started {
		return errors.Newf("operator already started")
	}

	for _, registeredStarter := range o.starters {
		if registeredStarter == starter {
			return errors.Newf("starter already registered")
		}
	}

	o.starters = append(o.starters, starter)

	return nil
}

func (o *operator) EnqueueItem(item operation.Item) {
	o.workqueue.Add(item.String())
}

func (o *operator) RegisterInformer(informer cache.SharedIndexInformer, group, version, kind string) error {
	o.lock.Lock()
	defer o.lock.Unlock()

	if o.started {
		return errors.Newf("operator already started")
	}

	for _, registeredInformer := range o.informers {
		if registeredInformer == informer {
			return errors.Newf("informer already registered")
		}
	}

	o.informers = append(o.informers, informer)

	informer.AddEventHandler(newResourceEventHandler(o, group, version, kind))

	return nil
}

func (o *operator) GetLogger() *zerolog.Logger {
	return &o.logger
}

func (o *operator) Start(threadiness int, stopCh <-chan struct{}) error {
	o.lock.Lock()
	defer o.lock.Unlock()

	if o.started {
		return errors.Newf("operator already started")
	}

	o.started = true

	return o.start(threadiness, stopCh)
}

func (o *operator) start(threadiness int, stopCh <-chan struct{}) error {
	// Execute pre checks
	o.logger.Info().Msgf("Executing Lifecycle PreStart")
	for _, handler := range o.handlers {
		if err := ExecLifecyclePreStart(handler); err != nil {
			return err
		}
	}

	o.logger.Info().Msgf("Starting informers")
	for _, starter := range o.starters {
		starter.Start(stopCh)
	}

	if err := o.waitForCacheSync(stopCh); err != nil {
		return err
	}

	o.logger.Info().Msgf("Starting workers")
	for i := 0; i < threadiness; i++ {
		go wait.Until(o.worker, time.Second, stopCh)
	}

	o.logger.Info().Msgf("Operator started")
	return nil
}

func (o *operator) waitForCacheSync(stopCh <-chan struct{}) error {
	cacheSync := make([]cache.InformerSynced, len(o.informers))

	for id, informer := range o.informers {
		cacheSync[id] = informer.HasSynced
	}

	if ok := cache.WaitForCacheSync(stopCh, cacheSync...); !ok {
		return errors.Newf("cache can not sync")
	}

	return nil
}
