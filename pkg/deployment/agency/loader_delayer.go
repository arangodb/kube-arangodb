//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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

package agency

import (
	"context"
	"sync"
	"time"

	agencyCache "github.com/arangodb/kube-arangodb/pkg/deployment/agency/cache"
)

func DelayLoader[T interface{}](loader agencyCache.StateLoader[T], delay time.Duration) agencyCache.StateLoader[T] {
	if delay <= 0 {
		return loader
	}

	return &delayerLoader[T]{
		parent: loader,
		delay:  delay,
	}
}

type delayerLoader[T interface{}] struct {
	lock sync.Mutex

	last  time.Time
	delay time.Duration

	parent agencyCache.StateLoader[T]
}

func (i *delayerLoader[T]) UpdateTime() time.Time {
	i.lock.Lock()
	defer i.lock.Unlock()

	return i.parent.UpdateTime()
}

func (i *delayerLoader[T]) Valid() bool {
	i.lock.Lock()
	defer i.lock.Unlock()

	return i.parent.Valid()
}

func (i *delayerLoader[T]) State() (*T, uint64, bool) {
	i.lock.Lock()
	defer i.lock.Unlock()

	return i.parent.State()
}

func (i *delayerLoader[T]) Invalidate() {
	i.lock.Lock()
	defer i.lock.Unlock()

	i.parent.Invalidate()
}

func (i *delayerLoader[T]) Refresh(ctx context.Context, discovery agencyCache.LeaderDiscovery) error {
	i.lock.Lock()
	defer i.lock.Unlock()

	if !i.parent.Valid() || i.last.IsZero() || time.Since(i.last) > i.delay {
		if err := i.parent.Refresh(ctx, discovery); err != nil {
			return err
		}

		i.last = time.Now()
	}

	return nil
}
