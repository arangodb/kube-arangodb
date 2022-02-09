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

package reconcile

import (
	"context"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
	"github.com/rs/zerolog"
)

// createNormalPlan considers the given specification & status and creates a plan to get the status in line with the specification.
// If a plan already exists, the given plan is returned with false.
// Otherwise the new plan is returned with a boolean true.
func createNormalPlan(ctx context.Context, log zerolog.Logger, apiObject k8sutil.APIObject,
	currentPlan api.Plan, spec api.DeploymentSpec,
	status api.DeploymentStatus, cachedStatus inspectorInterface.Inspector,
	builderCtx PlanBuilderContext) (api.Plan, api.BackOff, bool) {
	if !currentPlan.IsEmpty() {
		// Plan already exists, complete that first
		return currentPlan, nil, false
	}

	r := recoverPlanAppender(log, newPlanAppender(NewWithPlanBuilder(ctx, log, apiObject, spec, status, cachedStatus, builderCtx), status.BackOff, currentPlan).
		// Adjust topology settings
		ApplyIfEmpty(createTopologyMemberAdjustmentPlan).
		// Define topology
		ApplyIfEmpty(createTopologyEnablementPlan).
		ApplyIfEmpty(createTopologyUpdatePlan).
		// Check for scale up
		ApplyIfEmpty(createScaleUPMemberPlan).
		// Check for failed members
		ApplyIfEmpty(createMemberFailedRestorePlan).
		// Check for scale up/down
		ApplyIfEmpty(createScaleMemberPlan).
		// Update status
		ApplySubPlanIfEmpty(createEncryptionKeyStatusPropagatedFieldUpdate, createEncryptionKeyStatusUpdate).
		ApplyIfEmpty(createTLSStatusUpdate).
		ApplyIfEmpty(createJWTStatusUpdate).
		// Check for cleaned out dbserver in created state
		ApplyIfEmpty(createRemoveCleanedDBServersPlan).
		// Check for members to be removed
		ApplyIfEmpty(createReplaceMemberPlan).
		// Check for the need to rotate one or more members
		ApplyIfEmpty(createMarkToRemovePlan).
		ApplyIfEmpty(createRotateOrUpgradePlan).
		// Disable maintenance if upgrade process was done. Upgrade task throw IDLE Action if upgrade is pending
		ApplyIfEmpty(createMaintenanceManagementPlan).
		// Add keys
		ApplySubPlanIfEmpty(createEncryptionKeyStatusPropagatedFieldUpdate, createEncryptionKey).
		ApplyIfEmpty(createJWTKeyUpdate).
		ApplySubPlanIfEmpty(createTLSStatusPropagatedFieldUpdate, createCARenewalPlan).
		ApplySubPlanIfEmpty(createTLSStatusPropagatedFieldUpdate, createCAAppendPlan).
		ApplyIfEmpty(createKeyfileRenewalPlan).
		ApplyIfEmpty(createRotateServerStorageResizePlan).
		ApplySubPlanIfEmpty(createTLSStatusPropagatedFieldUpdate, createRotateTLSServerSNIPlan).
		ApplyIfEmpty(createRestorePlan).
		ApplySubPlanIfEmpty(createEncryptionKeyStatusPropagatedFieldUpdate, createEncryptionKeyCleanPlan).
		ApplySubPlanIfEmpty(createTLSStatusPropagatedFieldUpdate, createCACleanPlan).
		ApplyIfEmpty(createClusterOperationPlan).
		ApplyIfEmpty(createRebalancerGeneratePlan).
		// Final
		ApplyIfEmpty(createTLSStatusPropagated).
		ApplyIfEmpty(createBootstrapPlan))

	return r.Plan(), r.BackOff(), true
}

func createMemberFailedRestorePlan(ctx context.Context,
	log zerolog.Logger, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	cachedStatus inspectorInterface.Inspector, context PlanBuilderContext) api.Plan {
	var plan api.Plan

	// Fetch agency plan
	agencyState, agencyOK := context.GetAgencyCache()

	// Check for members in failed state
	status.Members.ForeachServerGroup(func(group api.ServerGroup, members api.MemberStatusList) error {
		for _, m := range members {
			if m.Phase != api.MemberPhaseFailed || len(plan) > 0 {
				continue
			}

			memberLog := log.Info().Str("id", m.ID).Str("role", group.AsRole())

			if group == api.ServerGroupDBServers && spec.GetMode() == api.DeploymentModeCluster {
				// Do pre check for DBServers. If agency is down DBServers should not be touch
				if !agencyOK {
					memberLog.Msg("Agency state is not present")
					continue
				}

				if agencyState.Plan.Collections.IsDBServerInDatabases(m.ID) {
					// DBServer still exists in agency plan! Will not be removed, but needs to be recreated
					memberLog.Msg("Recreating DBServer - it cannot be removed gracefully")
					plan = append(plan,
						api.NewAction(api.ActionTypeRecreateMember, group, m.ID))
					continue
				}

				// Everything is fine, proceed
			}

			switch group {
			case api.ServerGroupAgents:
				// For agents just recreate member do not rotate ID, do not remove PVC or service
				memberLog.Msg("Restoring old member. For agency members recreation of PVC is not supported - to prevent DataLoss")
				plan = append(plan,
					api.NewAction(api.ActionTypeRecreateMember, group, m.ID))
			case api.ServerGroupSingle:
				// Do not remove data for singles
				memberLog.Msg("Restoring old member. Rotation for single servers is not safe")
				plan = append(plan,
					api.NewAction(api.ActionTypeRecreateMember, group, m.ID))
			default:
				if spec.GetAllowMemberRecreation(group) {
					memberLog.Msg("Creating member replacement plan because member has failed")
					plan = append(plan,
						api.NewAction(api.ActionTypeRemoveMember, group, m.ID),
						api.NewAction(api.ActionTypeAddMember, group, ""),
						api.NewAction(api.ActionTypeWaitForMemberUp, group, api.MemberIDPreviousAction),
					)
				} else {
					memberLog.Msg("Restoring old member. Recreation is disabled for group")
					plan = append(plan,
						api.NewAction(api.ActionTypeRecreateMember, group, m.ID))
				}
			}
		}
		return nil
	})

	// Ensure that we were able to get agency info
	if len(plan) == 0 && !agencyOK {
		log.Warn().Msgf("unable to build further plan without access to agency")
		plan = append(plan,
			api.NewAction(api.ActionTypeIdle, api.ServerGroupUnknown, ""))
	}

	return plan
}

func createRemoveCleanedDBServersPlan(ctx context.Context,
	log zerolog.Logger, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	cachedStatus inspectorInterface.Inspector, context PlanBuilderContext) api.Plan {
	for _, m := range status.Members.DBServers {
		if !m.Phase.IsReady() {
			// Ensure that we CleanOut members which are Ready only to ensure data will be moved
			continue
		}

		if m.Phase.IsCreatedOrDrain() && m.Conditions.IsTrue(api.ConditionTypeCleanedOut) {
			log.Debug().
				Str("id", m.ID).
				Str("role", api.ServerGroupDBServers.AsRole()).
				Msg("Creating dbserver replacement plan because server is cleanout in created phase")
			return cleanOutMember(api.ServerGroupDBServers, m)
		}
	}

	return nil
}
