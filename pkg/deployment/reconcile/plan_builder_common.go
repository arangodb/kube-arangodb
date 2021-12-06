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
// Author Adam Janikowski
// Author Tomasz Mielech
//

package reconcile

import (
	"context"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
	"github.com/rs/zerolog"
)

func createMaintenanceManagementPlan(ctx context.Context,
	log zerolog.Logger, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	cachedStatus inspectorInterface.Inspector, planCtx PlanBuilderContext) api.Plan {
	if spec.Mode.Get() == api.DeploymentModeSingle {
		return nil
	}

	if !features.Maintenance().Enabled() {
		// Maintenance feature is not enabled
		return nil
	}

	agencyState, agencyOK := planCtx.GetAgencyCache()
	if !agencyOK {
		log.Error().Msgf("Unable to get agency mode")
		return nil
	}

	enabled := agencyState.Supervision.Maintenance.Exists()

	if !enabled && spec.Database.GetMaintenance() {
		log.Info().Msgf("Enabling maintenance mode")
		return api.Plan{api.NewAction(api.ActionTypeEnableMaintenance, api.ServerGroupUnknown, ""), api.NewAction(api.ActionTypeSetMaintenanceCondition, api.ServerGroupUnknown, "")}
	}

	if enabled && !spec.Database.GetMaintenance() {
		log.Info().Msgf("Disabling maintenance mode")
		return api.Plan{api.NewAction(api.ActionTypeDisableMaintenance, api.ServerGroupUnknown, ""), api.NewAction(api.ActionTypeSetMaintenanceCondition, api.ServerGroupUnknown, "")}
	}

	condition, ok := status.Conditions.Get(api.ConditionTypeMaintenanceMode)

	if enabled != (ok && condition.IsTrue()) {
		return api.Plan{api.NewAction(api.ActionTypeSetMaintenanceCondition, api.ServerGroupUnknown, "")}
	}

	return nil
}
