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

	"github.com/arangodb/kube-arangodb/pkg/deployment/agency/state"
)

func RefreshLoader(loader StateLoader, delay time.Duration) StateLoader {
	if delay <= 0 {
		return loader
	}

	return &refresherLoader{
		parent: loader,
		delay:  delay,
	}
}

type refresherLoader struct {
	lock sync.Mutex

	last  time.Time
	delay time.Duration

	parent StateLoader
}

func (i *refresherLoader) UpdateTime() time.Time {
	i.lock.Lock()
	defer i.lock.Unlock()

	return i.parent.UpdateTime()
}

func (i *refresherLoader) Valid() bool {
	i.lock.Lock()
	defer i.lock.Unlock()

	return i.parent.Valid()
}

func (i *refresherLoader) State() (*state.Root, uint64, bool) {
	i.lock.Lock()
	defer i.lock.Unlock()

	return i.parent.State()
}

func (i *refresherLoader) Invalidate() {
	i.lock.Lock()
	defer i.lock.Unlock()

	i.parent.Invalidate()
}

func (i *refresherLoader) Refresh(ctx context.Context, discovery LeaderDiscovery) error {
	i.lock.Lock()
	defer i.lock.Unlock()

	if i.last.IsZero() || time.Since(i.last) > i.delay {
		i.last = time.Now()
		i.parent.Invalidate()
	}

	if err := i.parent.Refresh(ctx, discovery); err != nil {
		return err
	}

	return nil
}
