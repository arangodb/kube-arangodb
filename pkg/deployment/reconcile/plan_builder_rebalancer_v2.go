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
	"time"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
)

func (r *Reconciler) createRebalancerV2GeneratePlan(spec api.DeploymentSpec, status api.DeploymentStatus) api.Plan {
	if spec.Mode.Get() != api.DeploymentModeCluster {
		return nil
	}

	if !spec.Rebalancer.IsEnabled() {
		r.metrics.Rebalancer.SetEnabled(false)

		if status.Rebalancer != nil {
			return api.Plan{
				api.NewAction(api.ActionTypeRebalancerCleanV2, api.ServerGroupUnknown, ""),
			}
		}
		return nil
	}

	r.metrics.Rebalancer.SetEnabled(true)

	if !status.Members.AllMembersReady(spec.Mode.Get(), spec.Sync.IsEnabled()) {
		return nil
	}

	if status.Rebalancer != nil {
		r.metrics.Rebalancer.SetCurrent(len(status.Rebalancer.MoveJobs))

		if len(status.Rebalancer.MoveJobs) != 0 {
			return nil
		}

		if !status.Rebalancer.LastCheckTime.IsZero() {
			if status.Rebalancer.LastCheckTime.Add(time.Minute).After(time.Now().UTC()) {
				return nil
			}
		}

		return api.Plan{
			api.NewAction(api.ActionTypeRebalancerGenerateV2, api.ServerGroupUnknown, ""),
		}
	} else {
		r.metrics.Rebalancer.current = 0
	}

	if status.Rebalancer == nil {
		return api.Plan{
			api.NewAction(api.ActionTypeRebalancerGenerateV2, api.ServerGroupUnknown, ""),
		}
	}

	return nil
}

func (r *Reconciler) createRebalancerV2CheckPlan(spec api.DeploymentSpec, status api.DeploymentStatus) api.Plan {
	if spec.Mode.Get() != api.DeploymentModeCluster {
		return nil
	}

	if status.Rebalancer == nil {
		return nil
	}

	if !status.Rebalancer.LastCheckTime.IsZero() && status.Rebalancer.LastCheckTime.Time.Add(5*time.Second).After(time.Now().UTC()) {
		return nil
	}
	if len(status.Rebalancer.MoveJobs) == 0 {
		return nil
	}

	return api.Plan{
		// Add plan to run check
		api.NewAction(api.ActionTypeRebalancerCheckV2, api.ServerGroupUnknown, ""),
	}
}
