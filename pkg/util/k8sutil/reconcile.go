//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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

package k8sutil

import (
	"context"
	"sync"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/refresh"
)

func NewReconcile(refresh refresh.Inspector) Reconcile {
	return &reconcile{refresh: refresh}
}

type Reconcile interface {
	Reconcile(ctx context.Context) error
	Required()
	IsRequired() bool
	WithError(err error) error

	ParallelAll(items int, executor func(id int) error) error
	Parallel(items, max int, executor func(id int) error) error
}

type reconcile struct {
	required bool

	refresh refresh.Inspector
}

func (r *reconcile) ParallelAll(items int, executor func(id int) error) error {
	return r.Parallel(items, items, executor)
}

func (r *reconcile) Parallel(items, max int, executor func(id int) error) error {
	var wg sync.WaitGroup

	l := make([]error, items)
	c := make(chan int, max)
	defer func() {
		close(c)
		for range c {

		}
	}()

	for i := 0; i < max; i++ {
		c <- 0
	}

	for i := 0; i < items; i++ {
		wg.Add(1)

		go func(id int) {
			defer wg.Done()

			defer func() {
				c <- 0
			}()

			<-c

			l[id] = executor(id)
		}(i)
	}

	wg.Wait()

	for i := 0; i < items; i++ {
		if l[i] == nil {
			continue
		}

		if errors.IsReconcile(l[i]) {
			continue
		}

		return l[i]
	}

	return nil
}

func (r *reconcile) WithRefresh(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}

	if errors.IsReconcile(err) {
		if r.refresh != nil {
			return r.refresh.Refresh(ctx)
		}

		return nil
	}

	return err
}

func (r *reconcile) Reconcile(ctx context.Context) error {
	if r.required {
		if err := r.refresh.Refresh(ctx); err != nil {
			return err
		}

		r.required = false
		return nil
	}

	return nil
}

func (r *reconcile) Required() {
	r.required = true
}

func (r *reconcile) IsRequired() bool {
	return r.required
}

func (r *reconcile) WithError(err error) error {
	if err == nil {
		return nil
	}

	if errors.IsReconcile(err) {
		r.Required()
		return nil
	}

	return err
}
