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
	"github.com/arangodb/kube-arangodb/pkg/deployment/reconcile/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

func (r *Reconciler) createSyncPlan(ctx context.Context, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	context PlanBuilderContext) api.Plan {

	if spec.Sync.IsEnabled() {
		if status.Conditions.IsTrue(api.ConditionTypeSyncEnabled) {
			// We do not have anything to do here
			return nil
		}

		// We need to enable condition
		return api.Plan{shared.UpdateConditionActionV2("Sync Enabled", api.ConditionTypeSyncEnabled, true, "Sync enabled", "", "")}
	}

	if !status.Conditions.IsTrue(api.ConditionTypeSyncEnabled) {
		// Sync is disabled
		return nil
	}

	cache, ok := context.GetAgencyArangoDBCache()
	if !ok {
		// Unable to get agency cache
		return nil
	}

	if cache.ArangoSync.Error == nil {
		if !cache.ArangoSync.IsSyncInProgress() {
			// Remove condition
			return api.Plan{shared.RemoveConditionActionV2("Sync Disabled", api.ConditionTypeSyncEnabled)}
		}
	}

	return nil
}
