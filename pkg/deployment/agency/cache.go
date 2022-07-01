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

package agency

import (
	"context"
	"sync"

	"github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/agency"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/generated/metric_descriptions"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/metrics"
)

type health struct {
	namespace, name string

	leaderID string
	leader   driver.Connection

	agencySize int

	names         []string
	commitIndexes map[string]uint64
	leaders       map[string]string
	election      map[string]int
}

func (h health) Leader() (driver.Connection, bool) {
	if l := h.leader; l != nil {
		return l, true
	}

	return nil, false
}

func (h health) CollectMetrics(m metrics.PushMetric) {
	if err := h.Serving(); err == nil {
		m.Push(metric_descriptions.ArangodbOperatorAgencyCacheServingGauge(1, h.namespace, h.name))
	} else {
		m.Push(metric_descriptions.ArangodbOperatorAgencyCacheServingGauge(0, h.namespace, h.name))
	}

	if err := h.Healthy(); err == nil {
		m.Push(metric_descriptions.ArangodbOperatorAgencyCacheHealthyGauge(1, h.namespace, h.name))
	} else {
		m.Push(metric_descriptions.ArangodbOperatorAgencyCacheHealthyGauge(0, h.namespace, h.name))
	}

	for _, name := range h.names {
		if i, ok := h.commitIndexes[name]; ok {
			m.Push(metric_descriptions.ArangodbOperatorAgencyCacheMemberServingGauge(1, h.namespace, h.name, name),
				metric_descriptions.ArangodbOperatorAgencyCacheMemberCommitOffsetGauge(float64(i), h.namespace, h.name, name))
		} else {
			m.Push(metric_descriptions.ArangodbOperatorAgencyCacheMemberServingGauge(0, h.namespace, h.name, name),
				metric_descriptions.ArangodbOperatorAgencyCacheMemberCommitOffsetGauge(-1, h.namespace, h.name, name))
		}
	}

	for k, l := range h.election {
		m.Push(metric_descriptions.ArangodbOperatorAgencyCacheLeadersGauge(float64(l), h.namespace, h.name, k))
	}
}

func (h health) LeaderID() string {
	return h.leaderID
}

// Healthy returns nil if all agencies have the same commit index.
func (h health) Healthy() error {
	if err := h.Serving(); err != nil {
		return err
	}

	if h.election[h.leaderID] != h.agencySize {
		return errors.Newf("Not all agents are in quorum")
	}

	index := h.commitIndexes[h.leaderID]
	if index == 0 {
		return errors.Newf("Agency CommitIndex is zero")
	}

	for k, v := range h.commitIndexes {
		if v != index {
			return errors.Newf("Agent %s is behind in CommitIndex", k)
		}
	}

	return nil
}

func (h health) Serving() error {
	if h.agencySize == 0 {
		return errors.Newf("Empty agents list")
	}

	if len(h.election) == 0 {
		return errors.Newf("No Leader")
	} else if len(h.election) > 1 {
		return errors.Newf("Multiple leaders")
	}

	if len(h.leaders) <= h.agencySize/2 {
		return errors.Newf("Quorum is not present")
	}

	return nil
}

// Health describes interface to check healthy of the environment.
type Health interface {
	// Healthy return nil when environment is considered as healthy.
	Healthy() error

	// Serving return nil when environment is considered as responsive, but not fully healthy.
	Serving() error

	// LeaderID returns a leader ID or empty string if a leader is not known.
	LeaderID() string

	// Leader returns connection to the Agency leader
	Leader() (driver.Connection, bool)

	CollectMetrics(m metrics.PushMetric)
}

type Cache interface {
	Reload(ctx context.Context, size int, clients map[string]agency.Agency) (uint64, error)
	Data() (State, bool)
	CommitIndex() uint64
	// Health returns true when healthy object is available.
	Health() (Health, bool)
}

func NewCache(namespace, name string, mode *api.DeploymentMode) Cache {
	if mode.Get() == api.DeploymentModeSingle {
		return NewSingleCache()
	}

	return NewAgencyCache(namespace, name)
}

func NewAgencyCache(namespace, name string) Cache {
	return &cache{
		namespace: namespace,
		name:      name,
	}
}

func NewSingleCache() Cache {
	return &cacheSingle{}
}

type cacheSingle struct {
}

func (c cacheSingle) CommitIndex() uint64 {
	return 0
}

// Health returns always false for single cache.
func (c cacheSingle) Health() (Health, bool) {
	return nil, false
}

func (c cacheSingle) Reload(_ context.Context, _ int, _ map[string]agency.Agency) (uint64, error) {
	return 0, nil
}

func (c cacheSingle) Data() (State, bool) {
	return State{}, true
}

type cache struct {
	namespace, name string

	lock sync.RWMutex

	valid bool

	commitIndex uint64

	data State

	health Health
}

func (c *cache) CommitIndex() uint64 {
	c.lock.RLock()
	defer c.lock.RUnlock()

	return c.commitIndex
}

func (c *cache) Data() (State, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	return c.data, c.valid
}

// Health returns always false for single cache.
func (c *cache) Health() (Health, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	if c.health != nil {
		return c.health, true
	}

	return nil, false
}

func (c *cache) Reload(ctx context.Context, size int, clients map[string]agency.Agency) (uint64, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	leaderCli, leaderConfig, health, err := getLeader(ctx, size, clients)
	if err != nil {
		// Invalidate a leader ID and agency state.
		// In the next iteration leaderID will be sat because `valid` will be false.
		c.valid = false
		c.health = nil

		return 0, err
	}

	health.namespace = c.namespace
	health.name = c.name

	c.health = health
	if leaderConfig.CommitIndex == c.commitIndex && c.valid {
		// We are on same index, nothing to do
		return leaderConfig.CommitIndex, nil
	}

	// A leader should be known even if an agency state is invalid.
	if data, err := loadState(ctx, leaderCli); err != nil {
		c.valid = false
		return leaderConfig.CommitIndex, err
	} else {
		c.data = data
		c.valid = true
		c.commitIndex = leaderConfig.CommitIndex
		return leaderConfig.CommitIndex, nil
	}
}

// getLeader returns config and client to a leader agency, and health to check if agencies are on the same page.
// If there is no quorum for the leader then error is returned.
func getLeader(ctx context.Context, size int, clients map[string]agency.Agency) (agency.Agency, *Config, health, error) {
	configs := make([]*Config, len(clients))
	errs := make([]error, len(clients))
	names := make([]string, 0, len(clients))
	for k := range clients {
		names = append(names, k)
	}

	var wg sync.WaitGroup

	// Fetch Agency config
	for i := range names {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			ctxLocal, cancel := globals.GetGlobals().Timeouts().Agency().WithTimeout(ctx)
			defer cancel()
			config, err := GetAgencyConfig(ctxLocal, clients[names[id]])

			if err != nil {
				errs[id] = err
				return
			}

			// Write config on the same index where client is (It will be helpful later).
			configs[id] = config
		}(i)
	}
	wg.Wait()

	var h health

	h.agencySize = size
	h.names = names
	h.commitIndexes = make(map[string]uint64, len(clients))
	h.leaders = make(map[string]string, len(clients))
	h.election = make(map[string]int, len(clients))

	for id := range configs {
		if config := configs[id]; config != nil {
			name := config.Configuration.ID
			if name == h.names[id] {
				h.commitIndexes[name] = config.CommitIndex
				if config.LeaderId != "" {
					h.leaders[name] = config.LeaderId
					h.election[config.LeaderId]++
					h.leaderID = config.LeaderId
				}
			}
		}
	}
	if err := h.Serving(); err != nil {
		logger.Err(err).Warn("Agency Not serving")
		return nil, nil, h, err
	}
	if err := h.Healthy(); err != nil {
		logger.Err(err).Warn("Agency Not healthy")
	}

	for id := range names {
		if h.leaderID == h.names[id] {
			h.leader = clients[names[id]].Connection()
			if cfg := configs[id]; cfg != nil {
				return clients[names[id]], cfg, h, nil
			}
		}
	}

	return nil, nil, h, errors.Newf("Unable to find agent")
}
