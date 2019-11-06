package reconcile

import (
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/rs/zerolog"
)

func createScaleMemeberPlan(log zerolog.Logger, spec api.DeploymentSpec, status api.DeploymentStatus) api.Plan {

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
