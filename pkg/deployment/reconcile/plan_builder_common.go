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
// Author Adam Janikowski
//

package reconcile

import (
	"context"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/agency"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	"github.com/arangodb/kube-arangodb/pkg/deployment/resources/inspector"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/rs/zerolog"
)

func createMaintenanceManagementPlan(ctx context.Context,
	log zerolog.Logger, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	cachedStatus inspector.Inspector, context PlanBuilderContext) api.Plan {
	if spec.Mode.Get() == api.DeploymentModeSingle {
		return nil
	}

	if !features.Maintenance().Enabled() {
		// Maintenance feature is not enabled
		return nil
	}

	client, err := context.GetDatabaseClient(ctx)
	if err != nil {
		log.Error().Err(err).Msgf("Unable to get agency client")
		return nil
	}

	m, err := agency.GetMaintenanceMode(ctx, client)
	if err != nil {
		log.Error().Err(err).Msgf("Unable to get agency maintenance mode")
		return nil
	}

	if !m.Enabled() && spec.Database.GetMaintenance() {
		log.Info().Msgf("Enabling maintenance mode")
		return api.Plan{api.NewAction(api.ActionTypeEnableMaintenance, api.ServerGroupUnknown, "")}
	}

	if m.Enabled() && !spec.Database.GetMaintenance() {
		log.Info().Msgf("Disabling maintenance mode")
		return api.Plan{api.NewAction(api.ActionTypeEnableMaintenance, api.ServerGroupUnknown, "")}
	}

	return nil
}
