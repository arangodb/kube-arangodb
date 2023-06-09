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
)

func DelayLoader[T interface{}](loader StateLoader[T], delay time.Duration) StateLoader[T] {
	if delay <= 0 {
		return loader
	}

	return &delayerLoader[T]{
		StateLoader: loader,
		delay:       delay,
	}
}

type delayerLoader[T interface{}] struct {
	lock sync.Mutex

	last  time.Time
	delay time.Duration

	StateLoader[T]
}

func (i *delayerLoader[T]) Refresh(ctx context.Context, discovery LeaderDiscovery) error {
	i.lock.Lock()
	defer i.lock.Unlock()

	if !i.StateLoader.Valid() || i.last.IsZero() || time.Since(i.last) > i.delay {
		if err := i.StateLoader.Refresh(ctx, discovery); err != nil {
			return err
		}

		i.last = time.Now()
	}

	return nil
}
