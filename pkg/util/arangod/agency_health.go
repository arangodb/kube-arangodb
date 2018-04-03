//
// DISCLAIMER
//
// Copyright 2018 ArangoDB GmbH, Cologne, Germany
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
// Author Ewout Prangsma
//

package arangod

import (
	"context"
	"fmt"
	"sync"
	"time"
)

const (
	maxAgentResponseTime = time.Second * 10
)

// agentStatus is a helper structure used in AreAgentsHealthy.
type agentStatus struct {
	IsLeader       bool
	LeaderEndpoint string
	IsResponding   bool
}

// AreAgentsHealthy performs a health check on all given agents.
// Of the given agents, 1 must respond as leader and all others must redirect to the leader.
// The function returns nil when all agents are healthy or an error when something is wrong.
func AreAgentsHealthy(ctx context.Context, clients []Agency) error {
	wg := sync.WaitGroup{}
	invalidKey := []string{"does-not-exist-149e97e8-4b81-5664-a8a8-9ba93881d64c"}
	statuses := make([]agentStatus, len(clients))
	for i, c := range clients {
		wg.Add(1)
		go func(i int, c Agency) {
			defer wg.Done()
			var trash interface{}
			lctx, cancel := context.WithTimeout(ctx, maxAgentResponseTime)
			defer cancel()
			if err := c.ReadKey(lctx, invalidKey, &trash); err == nil || IsKeyNotFound(err) {
				// We got a valid read from the leader
				statuses[i].IsLeader = true
				statuses[i].LeaderEndpoint = c.Endpoint()
				statuses[i].IsResponding = true
			} else {
				if location, ok := IsNotLeader(err); ok {
					// Valid response from a follower
					statuses[i].IsLeader = false
					statuses[i].LeaderEndpoint = location
					statuses[i].IsResponding = true
				} else {
					// Unexpected / invalid response
					statuses[i].IsResponding = false
				}
			}
		}(i, c)
	}
	wg.Wait()

	// Check the results
	noLeaders := 0
	for i, status := range statuses {
		if !status.IsResponding {
			return maskAny(fmt.Errorf("Agent %s is not responding", clients[i].Endpoint()))
		}
		if status.IsLeader {
			noLeaders++
		}
		if i > 0 {
			// Compare leader endpoint with previous
			prev := statuses[i-1].LeaderEndpoint
			if !IsSameEndpoint(prev, status.LeaderEndpoint) {
				return maskAny(fmt.Errorf("Not all agents report the same leader endpoint"))
			}
		}
	}
	if noLeaders != 1 {
		return maskAny(fmt.Errorf("Unexpected number of agency leaders: %d", noLeaders))
	}
	return nil
}
