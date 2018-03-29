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

package resilience

import (
	"time"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
)

const (
	recentTerminationsSinceGracePeriod = time.Minute * 10
	recentTerminationThreshold         = 5
)

// CheckMemberFailure performs a check for members that should be in failed state because:
// - They are frequently restarted
// - They cannot be scheduled for a long time (TODO)
func (r *Resilience) CheckMemberFailure() error {
	status := r.context.GetStatus()
	updateStatusNeeded := false
	if err := status.Members.ForeachServerGroup(func(group api.ServerGroup, list *api.MemberStatusList) error {
		for _, m := range *list {
			log := r.log.With().
				Str("id", m.ID).
				Str("role", group.AsRole()).
				Logger()
			// Check current state
			if m.Phase != api.MemberPhaseCreated {
				continue
			}
			// Check if pod is ready
			if m.Conditions.IsTrue(api.ConditionTypeReady) {
				continue
			}
			// Check recent terminations
			count := m.RecentTerminationsSince(time.Now().Add(-recentTerminationsSinceGracePeriod))
			if count >= recentTerminationThreshold {
				// Member has terminated too often in recent history.
				failureAcceptable, reason, err := r.isMemberFailureAcceptable(status, group, m)
				if err != nil {
					log.Warn().Err(err).Msg("Failed to check is member failure is acceptable")
				} else if failureAcceptable {
					log.Info().Msg("Member has terminated too often in recent history, marking is failed")
					m.Phase = api.MemberPhaseFailed
					list.Update(m)
					updateStatusNeeded = true
				} else {
					log.Warn().Msgf("Member has terminated too often in recent history, but it is not safe to mark it a failed because: %s", reason)
				}
			}
		}

		return nil
	}); err != nil {
		return maskAny(err)
	}
	if updateStatusNeeded {
		if err := r.context.UpdateStatus(status); err != nil {
			return maskAny(err)
		}
	}

	return nil
}

// isMemberFailureAcceptable checks if it is currently acceptable to switch the phase of the given member
// to failed, which means that it will be replaced.
// Return: failureAcceptable, notAcceptableReason, error
func (r *Resilience) isMemberFailureAcceptable(status api.DeploymentStatus, group api.ServerGroup, m api.MemberStatus) (bool, string, error) {
	switch group {
	case api.ServerGroupCoordinators:
		return true, "", nil
	default:
		// TODO
		return false, "TODO", nil
	}
}
