//
// DISCLAIMER
//
// Copyright 2025 ArangoDB GmbH, Cologne, Germany
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

package executor

import (
	"context"
	"sync"

	"github.com/rs/zerolog"

	"github.com/arangodb/kube-arangodb/pkg/logging"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func Run(ctx context.Context, log logging.Logger, threads int, f RunFunc) error {
	h := &handler{
		th:        NewThreadManager(threads),
		completed: make(chan any),
		log:       log,
	}

	go h.run(ctx, f)

	return h.Wait()
}

type RunFunc func(ctx context.Context, log logging.Logger, t Thread, h Handler) error

type Executor interface {
	Completed() bool

	Wait() error
}

type Handler interface {
	RunAsync(ctx context.Context, f RunFunc) Executor

	WaitForSubThreads(t Thread)
}

type handler struct {
	lock sync.Mutex

	th ThreadManager

	handlers []*handler

	completed chan any

	log logging.Logger

	err error
}

func (h *handler) WaitForSubThreads(t Thread) {
	for {
		t.Release()

		if h.subThreadsCompleted() {
			return
		}
	}
}

func (h *handler) subThreadsCompleted() bool {
	h.lock.Lock()
	defer h.lock.Unlock()

	for id := range h.handlers {
		if !h.handlers[id].Completed() {
			return false
		}
	}

	return true
}

func (h *handler) Wait() error {
	<-h.completed

	return h.err
}

func (h *handler) Completed() bool {
	select {
	case <-h.completed:
		return true
	default:
		return false
	}
}

func (h *handler) RunAsync(ctx context.Context, f RunFunc) Executor {
	h.lock.Lock()
	defer h.lock.Unlock()

	n := &handler{
		th:        h.th,
		completed: make(chan any),
		log:       h.log,
	}

	h.handlers = append(h.handlers, n)

	go n.run(ctx, f)

	return n
}

func (h *handler) run(ctx context.Context, entry RunFunc) {
	defer close(h.completed)

	err := h.runE(ctx, entry)

	subErrors := make([]error, len(h.handlers))

	for id := range subErrors {
		subErrors[id] = h.handlers[id].Wait()
	}

	subError := errors.Errors(subErrors...)

	h.err = errors.Errors(err, subError)
}

func (h *handler) runE(ctx context.Context, entry RunFunc) error {
	t := h.th.Acquire()
	defer t.Release()

	log := h.log.Wrap(func(in *zerolog.Event) *zerolog.Event {
		return in.Int("thread", int(t.ID()))
	})

	return entry(ctx, log, t, h)
}
