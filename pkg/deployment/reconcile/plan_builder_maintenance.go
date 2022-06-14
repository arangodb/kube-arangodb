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
				removeConditionActionV2("Backup not in progress", api.ConditionTypeBackupInProgress),
			}
		}

		// Backup is in progress

		hash := maintenance.Hash()

		if !currentCondition.IsTrue() || currentCondition.Hash != hash {
			return api.Plan{
				updateConditionActionV2("Backup in progress", api.ConditionTypeBackupInProgress, true, "Backup In Progress", "", hash),
			}
		}

		return nil
	} else {
		if maintenance.Exists() {
			return api.Plan{
				updateConditionActionV2("Backup in progress", api.ConditionTypeBackupInProgress, true, "Backup In Progress", "", maintenance.Hash()),
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
				removeConditionActionV2("Maintenance Disabled", api.ConditionTypeMaintenance),
			}
		}

		// Backup is in progress

		hash := maintenance.Hash()

		if !currentCondition.IsTrue() || currentCondition.Hash != hash {
			return api.Plan{
				updateConditionActionV2("Maintenance Enabled", api.ConditionTypeMaintenance, true, "Maintenance Enabled", "", hash),
			}
		}

		return nil
	} else {
		if maintenance.Exists() {
			return api.Plan{
				updateConditionActionV2("Maintenance Enabled", api.ConditionTypeMaintenance, true, "Maintenance Enabled", "", maintenance.Hash()),
			}
		}

		return nil
	}
}
