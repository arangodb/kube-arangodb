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

package agency

import (
	"context"
	"sync"

	"github.com/arangodb/go-driver/agency"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
)

type Cache interface {
	Reload(ctx context.Context, client agency.Agency) (uint64, error)
	Data() (State, bool)
	CommitIndex() uint64
}

func NewCache(mode *api.DeploymentMode) Cache {
	if mode.Get() == api.DeploymentModeSingle {
		return NewSingleCache()
	}

	return NewAgencyCache()
}

func NewAgencyCache() Cache {
	return &cache{}
}

func NewSingleCache() Cache {
	return &cacheSingle{}
}

type cacheSingle struct {
}

func (c cacheSingle) CommitIndex() uint64 {
	return 0
}

func (c cacheSingle) Reload(ctx context.Context, client agency.Agency) (uint64, error) {
	return 0, nil
}

func (c cacheSingle) Data() (State, bool) {
	return State{}, true
}

type cache struct {
	lock sync.Mutex

	valid bool

	commitIndex uint64

	data State
}

func (c *cache) CommitIndex() uint64 {
	return c.commitIndex
}

func (c *cache) Data() (State, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.data, c.valid
}

func (c *cache) Reload(ctx context.Context, client agency.Agency) (uint64, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	cfg, err := getAgencyConfig(ctx, client)
	if err != nil {
		c.valid = false
		return 0, err
	}

	if cfg.CommitIndex == c.commitIndex && c.valid {
		// We are on same index, nothing to do
		return cfg.CommitIndex, err
	}

	if data, err := loadState(ctx, client); err != nil {
		c.valid = false
		return cfg.CommitIndex, err
	} else {
		c.data = data
		c.valid = true
		c.commitIndex = cfg.CommitIndex
		return cfg.CommitIndex, nil
	}
}
