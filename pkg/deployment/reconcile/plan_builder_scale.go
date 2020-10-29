//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
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
// Author Tomasz Mielech <tomasz@arangodb.com>
//

package reconcile

import (
	"context"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/resources/inspector"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/rs/zerolog"
)

func createScaleMemberPlan(ctx context.Context,
	log zerolog.Logger, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	cachedStatus inspector.Inspector, context PlanBuilderContext) api.Plan {

	var plan api.Plan

	switch spec.GetMode() {
	case api.DeploymentModeSingle:
		// Never scale down
	case api.DeploymentModeActiveFailover:
		// Only scale singles
		plan = append(plan, createScalePlan(log, status.Members.Single, api.ServerGroupSingle, spec.Single.GetCount())...)
	case api.DeploymentModeCluster:
		// Scale dbservers, coordinators
		plan = append(plan, createScalePlan(log, status.Members.DBServers, api.ServerGroupDBServers, spec.DBServers.GetCount())...)
		plan = append(plan, createScalePlan(log, status.Members.Coordinators, api.ServerGroupCoordinators, spec.Coordinators.GetCount())...)
	}
	if spec.GetMode().SupportsSync() {
		// Scale syncmasters & syncworkers
		plan = append(plan, createScalePlan(log, status.Members.SyncMasters, api.ServerGroupSyncMasters, spec.SyncMasters.GetCount())...)
		plan = append(plan, createScalePlan(log, status.Members.SyncWorkers, api.ServerGroupSyncWorkers, spec.SyncWorkers.GetCount())...)
	}

	return plan
}

// createScalePlan creates a scaling plan for a single server group
func createScalePlan(log zerolog.Logger, members api.MemberStatusList, group api.ServerGroup, count int) api.Plan {
	var plan api.Plan
	if len(members) < count {
		// Scale up
		toAdd := count - len(members)
		for i := 0; i < toAdd; i++ {
			plan = append(plan, api.NewAction(api.ActionTypeAddMember, group, ""))
		}
		log.Debug().
			Int("count", count).
			Int("actual-count", len(members)).
			Int("delta", toAdd).
			Str("role", group.AsRole()).
			Msg("Creating scale-up plan")
	} else if len(members) > count {
		// Note, we scale down 1 member at a time
		if m, err := members.SelectMemberToRemove(); err != nil {
			log.Warn().Err(err).Str("role", group.AsRole()).Msg("Failed to select member to remove")
		} else {

			log.Debug().
				Str("member-id", m.ID).
				Str("phase", string(m.Phase)).
				Msg("Found member to remove")
			if group == api.ServerGroupDBServers {
				plan = append(plan,
					api.NewAction(api.ActionTypeCleanOutMember, group, m.ID),
				)
			}
			plan = append(plan,
				api.NewAction(api.ActionTypeShutdownMember, group, m.ID),
				api.NewAction(api.ActionTypeRemoveMember, group, m.ID),
			)
			log.Debug().
				Int("count", count).
				Int("actual-count", len(members)).
				Str("role", group.AsRole()).
				Str("member-id", m.ID).
				Msg("Creating scale-down plan")
		}
	}
	return plan
}

func createReplaceMemberPlan(ctx context.Context,
	log zerolog.Logger, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	cachedStatus inspector.Inspector, context PlanBuilderContext) api.Plan {

	var plan api.Plan

	// Replace is only allowed for DBServers
	switch spec.GetMode() {
	case api.DeploymentModeCluster:
		status.Members.ForeachServerInGroups(func(group api.ServerGroup, list api.MemberStatusList) error {
			for _, member := range list {
				if !plan.IsEmpty() {
					return nil
				}
				if member.Conditions.IsTrue(api.ConditionTypeMarkedToRemove) {
					plan = append(plan, api.NewAction(api.ActionTypeAddMember, group, "").
						AddParam(api.ActionTypeWaitForMemberInSync.String(), "").
						AddParam(api.ActionTypeWaitForMemberUp.String(), ""))
					log.Debug().
						Str("role", group.AsRole()).
						Msg("Creating replacement plan")
					return nil
				}
			}

			return nil
		}, api.ServerGroupDBServers)
	}

	return plan
}
