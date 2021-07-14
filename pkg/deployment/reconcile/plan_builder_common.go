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
	"time"

	"github.com/arangodb/kube-arangodb/pkg/util/arangod"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/agency"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
	"github.com/rs/zerolog"
)

const (
	MaintenanceDuration    = time.Hour
	MaintenanceGracePeriod = MaintenanceDuration / 2
)

func refreshMaintenance(ctx context.Context,
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

	condition, ok := status.Conditions.Get(api.ConditionTypeMaintenanceMode)

	if !ok || !condition.IsTrue() {
		return nil
	}

	// Check GracePeriod
	if condition.LastUpdateTime.Add(MaintenanceGracePeriod).After(time.Now()) {
		if err := planCtx.WithStatusUpdate(ctx, func(s *api.DeploymentStatus) bool {
			return s.Conditions.Touch(api.ConditionTypeMaintenanceMode)
		}); err != nil {
			return nil
		}
	}

	return nil
}

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

	ctxChild, cancel := context.WithTimeout(ctx, arangod.GetRequestTimeout())
	defer cancel()
	client, err := planCtx.GetDatabaseClient(ctxChild)
	if err != nil {
		log.Error().Err(err).Msgf("Unable to get agency client")
		return nil
	}

	ctxChild, cancel = context.WithTimeout(ctx, arangod.GetRequestTimeout())
	defer cancel()
	m, err := agency.GetMaintenanceMode(ctxChild, client)
	if err != nil {
		log.Error().Err(err).Msgf("Unable to get agency maintenance mode")
		return nil
	}

	if !m.Enabled() && spec.Database.GetMaintenance() {
		log.Info().Msgf("Enabling maintenance mode")
		return api.Plan{api.NewAction(api.ActionTypeEnableMaintenance, api.ServerGroupUnknown, ""), api.NewAction(api.ActionTypeSetMaintenanceCondition, api.ServerGroupUnknown, "")}
	}

	if m.Enabled() && !spec.Database.GetMaintenance() {
		log.Info().Msgf("Disabling maintenance mode")
		return api.Plan{api.NewAction(api.ActionTypeDisableMaintenance, api.ServerGroupUnknown, ""), api.NewAction(api.ActionTypeSetMaintenanceCondition, api.ServerGroupUnknown, "")}
	}

	condition, ok := status.Conditions.Get(api.ConditionTypeMaintenanceMode)

	if m.Enabled() != (ok && condition.IsTrue()) {
		return api.Plan{api.NewAction(api.ActionTypeSetMaintenanceCondition, api.ServerGroupUnknown, "")}
	}

	return nil
}
