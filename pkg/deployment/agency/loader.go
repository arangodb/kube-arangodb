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

func getLoader[T interface{}]() agencyCache.StateLoader[T] {
	loader := getLoaderBase[T]()

	loader = InvalidateOnErrorLoader[T](loader)

	loader = DelayLoader[T](loader, agencyCache.GlobalConfig().RefreshDelay)
	loader = RefreshLoader[T](loader, agencyCache.GlobalConfig().RefreshInterval)

	return loader
}

type StateLoader[T interface{}] interface {
	State() (*T, uint64, bool)

	Invalidate()
	Valid() bool

	UpdateTime() time.Time

	Refresh(ctx context.Context, discovery agencyCache.LeaderDiscovery) error
}

func NewSimpleStateLoader[T interface{}]() agencyCache.StateLoader[T] {
	return &simpleStateLoader[T]{}
}

type simpleStateLoader[T interface{}] struct {
	lock sync.Mutex

	state *T
	index uint64
	valid bool

	updateTime time.Time
}

func (s *simpleStateLoader[T]) UpdateTime() time.Time {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.updateTime
}

func (s *simpleStateLoader[T]) Valid() bool {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.valid
}

func (s *simpleStateLoader[T]) State() (*T, uint64, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if !s.valid {
		return nil, 0, false
	}

	return s.state, s.index, true
}

func (s *simpleStateLoader[T]) Invalidate() {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.valid = false
}

func (s *simpleStateLoader[T]) Refresh(ctx context.Context, discovery agencyCache.LeaderDiscovery) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	conn, err := discovery.Discover(ctx)
	if err != nil {
		return err
	}

	cfg, err := GetAgencyConfig(ctx, conn)
	if err != nil {
		return err
	}

	if !s.valid || s.index != cfg.CommitIndex {
		// Full reload
		state, err := GetAgencyState[T](ctx, conn)
		if err != nil {
			return err
		}

		s.index = cfg.CommitIndex
		s.state = &state
		s.valid = true
		s.updateTime = time.Now()
	}

	return nil
}
