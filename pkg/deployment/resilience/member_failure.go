//
// DISCLAIMER
//
// Copyright 2020-2021 ArangoDB GmbH, Cologne, Germany
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
// Author Tomasz Mielech
//

package resilience

import (
	"context"
	"time"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"

	"github.com/arangodb/go-driver/agency"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod"
)

const (
	recentTerminationsSinceGracePeriod = time.Minute * 10
	notReadySinceGracePeriod           = time.Minute * 5
	recentTerminationThreshold         = 5
)

// CheckMemberFailure performs a check for members that should be in failed state because:
// - They are frequently restarted
// - They cannot be scheduled for a long time (TODO)
func (r *Resilience) CheckMemberFailure(ctx context.Context) error {
	status, lastVersion := r.context.GetStatus()
	updateStatusNeeded := false
	if err := status.Members.ForeachServerGroup(func(group api.ServerGroup, list api.MemberStatusList) error {
		for _, m := range list {
			log := r.log.With().
				Str("id", m.ID).
				Str("role", group.AsRole()).
				Logger()

			// Check if there are Members with Phase Upgrading or Rotation but no plan
			switch m.Phase {
			case api.MemberPhaseNone, api.MemberPhasePending:
				continue
			case api.MemberPhaseUpgrading, api.MemberPhaseRotating, api.MemberPhaseCleanOut, api.MemberPhaseRotateStart:
				if len(status.Plan) == 0 {
					log.Error().Msgf("No plan but member is in phase %s - marking as failed", m.Phase)
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

			// Check not ready for a long time
			if !m.Phase.IsFailed() {
				if m.IsNotReadySince(time.Now().Add(-notReadySinceGracePeriod)) {
					// Member has terminated too often in recent history.

					failureAcceptable, reason, err := r.isMemberFailureAcceptable(ctx, group, m)
					if err != nil {
						log.Warn().Err(err).Msg("Failed to check is member failure is acceptable")
					} else if failureAcceptable {
						log.Info().Msg("Member is not ready for long time, marking is failed")
						m.Phase = api.MemberPhaseFailed
						status.Members.Update(m, group)
						updateStatusNeeded = true
					} else {
						log.Warn().Msgf("Member is not ready for long time, but it is not safe to mark it a failed because: %s", reason)
					}
				}
			}

			// Check recent terminations
			if !m.Phase.IsFailed() {
				count := m.RecentTerminationsSince(time.Now().Add(-recentTerminationsSinceGracePeriod))
				if count >= recentTerminationThreshold {
					// Member has terminated too often in recent history.
					failureAcceptable, reason, err := r.isMemberFailureAcceptable(ctx, group, m)
					if err != nil {
						log.Warn().Err(err).Msg("Failed to check is member failure is acceptable")
					} else if failureAcceptable {
						log.Info().Msg("Member has terminated too often in recent history, marking is failed")
						m.Phase = api.MemberPhaseFailed
						status.Members.Update(m, group)
						updateStatusNeeded = true
					} else {
						log.Warn().Msgf("Member has terminated too often in recent history, but it is not safe to mark it a failed because: %s", reason)
					}
				}
			}
		}

		return nil
	}); err != nil {
		return errors.WithStack(err)
	}
	if updateStatusNeeded {
		if err := r.context.UpdateStatus(ctx, status, lastVersion); err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

// isMemberFailureAcceptable checks if it is currently acceptable to switch the phase of the given member
// to failed, which means that it will be replaced.
// Return: failureAcceptable, notAcceptableReason, error
func (r *Resilience) isMemberFailureAcceptable(ctx context.Context, group api.ServerGroup, m api.MemberStatus) (bool, string, error) {

	switch group {
	case api.ServerGroupAgents:
		// All good when remaining agents are health
		ctxChild, cancel := context.WithTimeout(ctx, arangod.GetRequestTimeout())
		defer cancel()
		clients, err := r.context.GetAgencyClients(ctxChild, func(id string) bool { return id != m.ID })
		if err != nil {
			return false, "", errors.WithStack(err)
		}
		if err := agency.AreAgentsHealthy(ctx, clients); err != nil {
			return false, err.Error(), nil
		}
		return true, "", nil
	case api.ServerGroupDBServers:
		ctxChild, cancel := context.WithTimeout(ctx, arangod.GetRequestTimeout())
		defer cancel()
		client, err := r.context.GetDatabaseClient(ctxChild)
		if err != nil {
			return false, "", errors.WithStack(err)
		}
		if err := arangod.IsDBServerEmpty(ctx, m.ID, client); err != nil {
			return false, err.Error(), nil
		}
		return true, "", nil
	case api.ServerGroupCoordinators:
		// Coordinators can be replaced at will
		return true, "", nil
	case api.ServerGroupSyncMasters, api.ServerGroupSyncWorkers:
		// Sync masters & workers can be replaced at will
		return true, "", nil
	default:
		// TODO
		return false, "TODO", nil
	}
}
