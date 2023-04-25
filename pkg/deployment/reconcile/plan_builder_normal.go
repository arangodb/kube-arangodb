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

package reconcile

import (
	"context"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

// createNormalPlan considers the given specification & status and creates a plan to get the status in line with the specification.
// If a plan already exists, the given plan is returned with false.
// Otherwise the new plan is returned with a boolean true.
func (r *Reconciler) createNormalPlan(ctx context.Context, apiObject k8sutil.APIObject,
	currentPlan api.Plan, spec api.DeploymentSpec,
	status api.DeploymentStatus,
	builderCtx PlanBuilderContext) (api.Plan, api.BackOff, bool) {
	if !currentPlan.IsEmpty() {
		// Plan already exists, complete that first
		return currentPlan, nil, false
	}

	q := recoverPlanAppender(r.log, newPlanAppender(NewWithPlanBuilder(ctx, apiObject, spec, status, builderCtx), status.BackOff, currentPlan).
		// Define topology
		ApplyIfEmpty(r.createTopologyEnablementPlan).
		// Adjust topology settings
		ApplyIfEmpty(r.createTopologyMemberAdjustmentPlan).
		ApplyIfEmpty(r.createTopologyUpdatePlan).
		// Check for scale up
		ApplyIfEmpty(r.createScaleUPMemberPlan).
		// Check for failed members
		ApplyIfEmpty(r.createMemberFailedRestoreNormalPlan).
		ApplyIfEmpty(r.createRebuildOutSyncedPlan).
		// Check for scale up/down
		ApplyIfEmpty(r.createScaleMemberPlan).
		// Update status
		ApplySubPlanIfEmpty(r.createEncryptionKeyStatusPropagatedFieldUpdate, r.createEncryptionKeyStatusUpdate).
		ApplyIfEmpty(r.createTLSStatusUpdate).
		ApplyIfEmpty(r.createJWTStatusUpdate).
		// Check for cleaned out dbserver in created state
		ApplyIfEmpty(r.createRemoveCleanedDBServersPlan).
		// Check for members to be removed
		ApplyIfEmpty(r.createReplaceMemberPlan).
		// Check for the need to rotate one or more members
		ApplyIfEmpty(r.createMarkToRemovePlan).
		ApplyIfEmpty(r.createMemberMaintenanceManagementPlan).
		ApplyIfEmpty(r.createRotateOrUpgradePlan).
		// Disable maintenance if upgrade process was done. Upgrade task throw IDLE Action if upgrade is pending
		ApplyIfEmpty(r.createMaintenanceManagementPlan).
		// Add keys
		ApplySubPlanIfEmpty(r.createEncryptionKeyStatusPropagatedFieldUpdate, r.createEncryptionKey).
		ApplyIfEmpty(r.createJWTKeyUpdate).
		ApplySubPlanIfEmpty(r.createTLSStatusPropagatedFieldUpdate, r.createCARenewalPlan).
		ApplySubPlanIfEmpty(r.createTLSStatusPropagatedFieldUpdate, r.createCAAppendPlan).
		ApplyIfEmpty(r.createKeyfileRenewalPlan).
		ApplyIfEmpty(r.createRotateServerStorageResizePlanRotate).
		ApplySubPlanIfEmpty(r.createTLSStatusPropagatedFieldUpdate, r.createRotateTLSServerSNIPlan).
		ApplyIfEmpty(r.createRestorePlan).
		ApplySubPlanIfEmpty(r.createEncryptionKeyStatusPropagatedFieldUpdate, r.createEncryptionKeyCleanPlan).
		ApplySubPlanIfEmpty(r.createTLSStatusPropagatedFieldUpdate, r.createCACleanPlan).
		ApplyIfEmpty(r.createClusterOperationPlan).
		ApplyIfEmpty(r.createRebalancerGeneratePlan).
		// Final
		ApplyIfEmpty(r.createTLSStatusPropagated).
		ApplyIfEmpty(r.createBootstrapPlan))

	return q.Plan(), q.BackOff(), true
}

func (r *Reconciler) createRemoveCleanedDBServersPlan(ctx context.Context, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	context PlanBuilderContext) api.Plan {
	for _, m := range status.Members.DBServers {
		if !m.Phase.IsReady() {
			// Ensure that we CleanOut members which are Ready only to ensure data will be moved
			continue
		}

		if m.Phase.IsCreatedOrDrain() && m.Conditions.IsTrue(api.ConditionTypeCleanedOut) {
			r.log.
				Str("id", m.ID).
				Str("role", api.ServerGroupDBServers.AsRole()).
				Debug("Creating dbserver replacement plan because server is cleanout in created phase")
			return cleanOutMember(api.ServerGroupDBServers, m)
		}
	}

	return nil
}
