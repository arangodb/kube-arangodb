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
	"fmt"
	"sync"

	"github.com/arangodb/go-driver/agency"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type Cache interface {
	Reload(ctx context.Context, clients []agency.Agency) (uint64, error)
	Data() (State, bool)
	GetLeaderID() string
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

// GetLeaderID returns always empty string for a single cache.
func (c cacheSingle) GetLeaderID() string {
	return ""
}

func (c cacheSingle) Reload(_ context.Context, _ []agency.Agency) (uint64, error) {
	return 0, nil
}

func (c cacheSingle) Data() (State, bool) {
	return State{}, true
}

type cache struct {
	lock sync.RWMutex

	valid bool

	commitIndex uint64

	data State

	leaderID string
}

func (c *cache) Data() (State, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	return c.data, c.valid
}

// GetLeaderID returns a leader ID or empty string if a leader is not known.
func (c *cache) GetLeaderID() string {
	return c.leaderID
}

func (c *cache) Reload(ctx context.Context, clients []agency.Agency) (uint64, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	leaderCli, leaderConfig, err := getLeader(ctx, clients)
	if err != nil {
		// Invalidate a leader ID and agency state.
		// In the next iteration leaderID will be sat because `valid` will be false.
		c.leaderID = ""
		c.valid = false

		return 0, err
	}

	if leaderConfig.CommitIndex == c.commitIndex && c.valid {
		// We are on same index, nothing to do
		return leaderConfig.CommitIndex, nil
	}

	// A leader should be known even if an agency state is invalid.
	c.leaderID = leaderConfig.LeaderId
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

// getLeader returns config and client to a leader agency.
// If there is no quorum for the leader then error is returned.
func getLeader(ctx context.Context, clients []agency.Agency) (agency.Agency, *agencyConfig, error) {
	var mutex sync.Mutex
	var anyError error
	var wg sync.WaitGroup

	cliLen := len(clients)
	if cliLen == 0 {
		return nil, nil, errors.New("empty list of agencies' clients")
	}
	configs := make([]*agencyConfig, cliLen)
	leaders := make(map[string]int)

	// Fetch all configs from agencies.
	wg.Add(cliLen)
	for i, cli := range clients {
		go func(iLocal int, cliLocal agency.Agency) {
			defer wg.Done()
			config, err := getAgencyConfig(ctx, cliLocal)

			mutex.Lock()
			defer mutex.Unlock()

			if err != nil {
				anyError = err
				return
			} else if config == nil || config.LeaderId == "" {
				anyError = fmt.Errorf("leader unknown for the agent %v", cliLocal.Connection().Endpoints())
				return
			}

			// Write config on the same index where client is (It will be helpful later).
			configs[iLocal] = config
			// Count leaders.
			leaders[config.LeaderId]++
		}(i, cli)
	}
	wg.Wait()

	if len(leaders) == 0 {
		return nil, nil, wrapError(anyError, "failed to get config from agencies")
	}

	// Find the leader ID which has the most votes from all agencies.
	maxVotes := 0
	var leaderID string
	for id, votes := range leaders {
		if votes > maxVotes {
			maxVotes = votes
			leaderID = id
		}
	}

	// Check if a leader has quorum from all possible agencies.
	if maxVotes <= cliLen/2 {
		message := fmt.Sprintf("no quorum for leader %s, votes %d of %d", leaderID, maxVotes, cliLen)
		return nil, nil, wrapError(anyError, message)
	}

	// From here on, a leader with quorum is known.
	for i, config := range configs {
		if config != nil && config.LeaderId == leaderID {
			return clients[i], config, nil
		}
	}

	return nil, nil, wrapError(anyError, "the leader is not responsive")
}

func wrapError(err error, message string) error {
	if err != nil {
		return errors.WithMessage(err, message)
	}

	return errors.New(message)
}
