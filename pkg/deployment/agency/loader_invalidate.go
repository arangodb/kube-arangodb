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

func InvalidateOnErrorLoader(loader StateLoader) StateLoader {
	return &invalidateOnErrorLoader{
		parent: loader,
	}
}

type invalidateOnErrorLoader struct {
	lock sync.Mutex

	parent StateLoader
}

func (i *invalidateOnErrorLoader) UpdateTime() time.Time {
	i.lock.Lock()
	defer i.lock.Unlock()

	return i.parent.UpdateTime()
}

func (i *invalidateOnErrorLoader) Valid() bool {
	i.lock.Lock()
	defer i.lock.Unlock()

	return i.parent.Valid()
}

func (i *invalidateOnErrorLoader) State() (*state.Root, uint64, bool) {
	i.lock.Lock()
	defer i.lock.Unlock()

	return i.parent.State()
}

func (i *invalidateOnErrorLoader) Invalidate() {
	i.lock.Lock()
	defer i.lock.Unlock()

	i.parent.Invalidate()
}

func (i *invalidateOnErrorLoader) Refresh(ctx context.Context, discovery LeaderDiscovery) (err error) {
	i.lock.Lock()
	defer i.lock.Unlock()

	defer func() {
		if err != nil {
			i.parent.Invalidate()
		}
	}()

	defer func() {
		if p := recover(); p != nil {
			i.parent.Invalidate()
			panic(p)
		}
	}()

	return i.parent.Refresh(ctx, discovery)
}
