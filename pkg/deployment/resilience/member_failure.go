//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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

package resilience

import (
	"context"
	"time"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/agency/state"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

const (
	recentTerminationsSinceGracePeriod = time.Minute * 10
	recentTerminationThreshold         = 5
)

// CheckMemberFailure performs a check for members that should be in failed state because:
// - They are frequently restarted
// - They cannot be scheduled for a long time (TODO)
func (r *Resilience) CheckMemberFailure(ctx context.Context) error {
	status := r.context.GetStatus()
	updateStatusNeeded := false

	for _, e := range status.Members.AsList() {
		m := e.Member
		group := e.Group
		log := r.log("member-failure").
			Str("id", m.ID).
			Str("role", group.AsRole())

		// Check if there are Members with Phase Upgrading or Rotation but no plan
		switch m.Phase {
		case api.MemberPhaseNone, api.MemberPhasePending, api.MemberPhaseCreationFailed:
			continue
		case api.MemberPhaseUpgrading, api.MemberPhaseRotating, api.MemberPhaseCleanOut, api.MemberPhaseRotateStart, api.MemberPhaseShuttingDown:
			if len(status.Plan) == 0 {
				log.Error("No plan but member is in phase %s - marking as failed", m.Phase)
				m.Phase = api.MemberPhaseFailed
				status.Members.Update(m, group)
				updateStatusNeeded = true
			}
		}

		// Check if pod is ready
		if m.Conditions.IsTrue(api.ConditionTypeReady) {
			// Pod is now ready, so we're not looking further
			continue
		}

		// Check recent terminations
		if !m.Phase.IsFailed() {
			count := m.RecentTerminationsSince(time.Now().Add(-recentTerminationsSinceGracePeriod))
			if count >= recentTerminationThreshold {
				// Member has terminated too often in recent history.
				failureAcceptable, reason := r.isMemberFailureAcceptable(group, m)
				if failureAcceptable {
					log.Info("Member has terminated too often in recent history, marking is failed")
					m.Phase = api.MemberPhaseFailed
					status.Members.Update(m, group)
					updateStatusNeeded = true
				} else {
					log.Warn("Member has terminated too often in recent history, but it is not safe to mark it a failed because: %s", reason)
				}
			}
		}
	}

	if updateStatusNeeded {
		if err := r.context.UpdateStatus(ctx, status); err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

// isMemberFailureAcceptable checks if it is currently acceptable to switch the phase of the given member
// to failed, which means that it will be replaced.
// Return: failureAcceptable, notAcceptableReason
func (r *Resilience) isMemberFailureAcceptable(group api.ServerGroup, m api.MemberStatus) (bool, string) {

	switch group {
	case api.ServerGroupAgents:
		agencyHealth, ok := r.context.GetAgencyHealth()
		if !ok {
			return false, "AgencyHealth is not present"
		}

		if err := agencyHealth.Healthy(); err != nil {
			return false, err.Error()
		}

		return true, ""
	case api.ServerGroupDBServers:
		agencyState, ok := r.context.GetAgencyCache()
		if !ok {
			return false, "AgencyHealth is not present"
		}

		if agencyState.Plan.Collections.IsDBServerPresent(state.Server(m.ID)) {
			return false, "DBServer still in Plan"
		}

		if agencyState.Current.Collections.IsDBServerPresent(state.Server(m.ID)) {
			return false, "DBServer still in Current"
		}

		return true, ""
	case api.ServerGroupCoordinators:
		// Coordinators can be replaced at will
		return true, ""
	case api.ServerGroupSyncMasters, api.ServerGroupSyncWorkers:
		// Sync masters & workers can be replaced at will
		return true, ""
	case api.ServerGroupSingle:
		return false, "ServerGroupSingle can not marked as a failed"
	default:
		// TODO
		return false, "TODO"
	}
}
