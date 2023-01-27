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
	"time"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/actions"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	"github.com/arangodb/kube-arangodb/pkg/deployment/reconcile/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

func (r *Reconciler) createBackupInProgressConditionPlan(ctx context.Context, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	context PlanBuilderContext) api.Plan {

	if spec.Mode.Get() != api.DeploymentModeCluster {
		return nil
	}

	cache, ok := context.GetAgencyCache()
	if !ok {
		return nil
	}

	currentCondition, currentConditionExists := status.Conditions.Get(api.ConditionTypeBackupInProgress)

	maintenance := cache.Target.HotBackup.Create

	if currentConditionExists {
		// Condition exists
		if !maintenance.Exists() {
			// Condition needs to be removed
			return api.Plan{
				shared.RemoveConditionActionV2("Backup not in progress", api.ConditionTypeBackupInProgress),
			}
		}

		// Backup is in progress

		hash := maintenance.Hash()

		if !currentCondition.IsTrue() || currentCondition.Hash != hash {
			return api.Plan{
				shared.UpdateConditionActionV2("Backup in progress", api.ConditionTypeBackupInProgress, true, "Backup In Progress", "", hash),
			}
		}

		return nil
	} else {
		if maintenance.Exists() {
			return api.Plan{
				shared.UpdateConditionActionV2("Backup in progress", api.ConditionTypeBackupInProgress, true, "Backup In Progress", "", maintenance.Hash()),
			}
		}

		return nil
	}
}

func (r *Reconciler) createMaintenanceConditionPlan(ctx context.Context, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	context PlanBuilderContext) api.Plan {

	if spec.Mode.Get() != api.DeploymentModeCluster {
		return nil
	}

	cache, ok := context.GetAgencyCache()
	if !ok {
		return nil
	}

	currentCondition, currentConditionExists := status.Conditions.Get(api.ConditionTypeMaintenance)

	maintenance := cache.Supervision.Maintenance

	if currentConditionExists {
		// Condition exists
		if !maintenance.Exists() {
			// Condition needs to be removed
			return api.Plan{
				shared.RemoveConditionActionV2("Maintenance Disabled", api.ConditionTypeMaintenance),
			}
		}

		// Backup is in progress

		hash := maintenance.Hash()

		if !currentCondition.IsTrue() || currentCondition.Hash != hash {
			return api.Plan{
				shared.UpdateConditionActionV2("Maintenance Enabled", api.ConditionTypeMaintenance, true, "Maintenance Enabled", "", hash),
			}
		}

		return nil
	} else {
		if maintenance.Exists() {
			return api.Plan{
				shared.UpdateConditionActionV2("Maintenance Enabled", api.ConditionTypeMaintenance, true, "Maintenance Enabled", "", maintenance.Hash()),
			}
		}

		return nil
	}
}

func (r *Reconciler) createMaintenanceManagementPlan(ctx context.Context, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	planCtx PlanBuilderContext) api.Plan {
	if spec.Mode.Get() == api.DeploymentModeSingle {
		return nil
	}

	if !features.Maintenance().Enabled() {
		// Maintenance feature is not enabled
		return nil
	}

	agencyState, agencyOK := planCtx.GetAgencyCache()
	if !agencyOK {
		r.log.Error("Unable to get agency mode")
		return nil
	}

	if agencyState.Target.HotBackup.Create.Exists() {
		r.log.Info("HotBackup in progress")
		return nil
	}

	enabled := agencyState.Supervision.Maintenance.Exists()
	c, cok := status.Conditions.Get(api.ConditionTypeMaintenance)

	if (cok && c.IsTrue()) != enabled {
		// Condition not yet propagated
		r.log.Info("Condition not yet propagated")
		return nil
	}

	if cok {
		if t := c.LastTransitionTime.Time; !t.IsZero() {
			if time.Since(t) < 5*time.Second {
				// Did not elapse 5 s
				return nil
			}
		}
	}

	if !enabled && spec.Database.GetMaintenance() {
		r.log.Info("Enabling maintenance mode")
		return api.Plan{actions.NewClusterAction(api.ActionTypeEnableMaintenance)}
	}

	if enabled && !spec.Database.GetMaintenance() {
		r.log.Info("Disabling maintenance mode")
		return api.Plan{actions.NewClusterAction(api.ActionTypeDisableMaintenance)}
	}

	return nil
}
