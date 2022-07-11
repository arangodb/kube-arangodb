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
	"time"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/actions"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

var (
	ObsoleteClusterConditions = []api.ConditionType{
		api.ConditionTypeMaintenanceMode,
	}
)

func (r *Reconciler) cleanupConditions(ctx context.Context, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	planCtx PlanBuilderContext) api.Plan {
	var p api.Plan

	for _, c := range ObsoleteClusterConditions {
		if _, ok := status.Conditions.Get(c); ok {
			p = append(p, removeConditionActionV2("Cleanup Condition", c))
		}
	}

	return p
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
