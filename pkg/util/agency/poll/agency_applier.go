//
// DISCLAIMER
//
// Copyright 2016-2026 ArangoDB GmbH, Cologne, Germany
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

package poll

import (
	"context"
	"sync"
	"time"

	"github.com/arangodb-helper/go-helper/pkg/arangod/agency/cache"
	"github.com/arangodb-helper/go-helper/pkg/errors"
)

func NewAgencyApplier[T interface{}](cfg ApplierConfig) cache.StateLoader[T] {
	return &agencyApplier[T]{
		applier: NewApplier[T](cfg),
	}
}

type agencyApplier[T interface{}] struct {
	lock  sync.Mutex
	valid bool
	// index is last commit index which is applied to the structure `T`.
	// If it is 0 then applier has not executed any actions yet.
	index      uint64
	updateTime time.Time
	applier    Applier[T]
}

func (a *agencyApplier[T]) State() (*T, uint64, bool) {
	a.lock.Lock()
	defer a.lock.Unlock()
	if !a.valid {
		return nil, 0, false
	}
	return a.applier.Get(), a.index, true
}
func (a *agencyApplier[T]) Invalidate() {
	a.lock.Lock()
	defer a.lock.Unlock()
	a.valid = false
}
func (a *agencyApplier[T]) Valid() bool {
	a.lock.Lock()
	defer a.lock.Unlock()
	return a.valid
}
func (a *agencyApplier[T]) UpdateTime() time.Time {
	a.lock.Lock()
	defer a.lock.Unlock()
	return a.updateTime
}
func (a *agencyApplier[T]) Refresh(ctx context.Context, discovery cache.LeaderDiscovery) (err error) {
	a.lock.Lock()
	defer a.lock.Unlock()
	defer func() {
		if err != nil {
			// Invalidate result
			a.valid = false
		}
	}()
	defer func() {
		// Catch panic
		if p := recover(); p != nil {
			err = errors.WithStack(errors.Newf("Panic recovery: %+v", p))
		}
	}()
	leader, err := discovery.Discover(ctx)
	if err != nil {
		return errors.WithMessage(err, "Discover failed")
	}
	// index must be 0 when:
	// - it is a first call to this agency applier, so a.index is 0.
	// - it is not valid.
	var index uint64
	if a.valid && a.index > 0 {
		index = a.index + 1
	}
	resp, err := GetAgencyPoll[T](ctx, leader, index, 0)
	if err != nil {
		return errors.WithMessage(err, "GetAgencyPoll failed")
	}
	if resp.Result.FirstIndex == nil {
		if resp.Result.CommitIndex != a.index {
			// Indexes are wrong!
			return errors.Newf("Invalid index")
		}
		// Nothing to do
		return nil
	} else if idx := *resp.Result.FirstIndex; idx == 0 {
		// Full reload
		if resp.Result.ReadDB == nil {
			// There is no data
			return errors.Newf("Data is missing")
		}
		a.applier.Set(resp.Result.ReadDB)
		a.index = resp.Result.CommitIndex
		a.valid = true
		return nil
	} else {
		// Index is set, we are parsing results
		for _, items := range resp.Result.Log {
			if err := a.applier.ApplyItemSet(items.Items); err != nil {
				return errors.WithMessage(err, "ApplyItemSet failed")
			}
		}
		a.valid = true
		a.index = resp.Result.CommitIndex
		return nil
	}
}
