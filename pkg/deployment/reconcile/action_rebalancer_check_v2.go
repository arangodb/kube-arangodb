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

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/agency/state"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

func newRebalancerCheckV2Action(action api.Action, actionCtx ActionContext) Action {
	a := &actionRebalancerCheckV2{}

	a.actionImpl = newActionImplDefRef(action, actionCtx)

	return a
}

type actionRebalancerCheckV2 struct {
	actionImpl

	actionEmptyCheckProgress
}

func (r actionRebalancerCheckV2) Start(ctx context.Context) (bool, error) {
	rebalancerStatus := r.actionCtx.GetStatus().Rebalancer

	if rebalancerStatus == nil {
		return true, nil
	}

	if len(rebalancerStatus.MoveJobs) == 0 {
		return true, nil
	}

	cache, ok := r.actionCtx.GetAgencyCache()
	if !ok {
		r.log.Debug("AgencyCache is not ready")
		return true, nil
	}

	statuses := make([]state.JobPhase, len(rebalancerStatus.MoveJobs))

	for id := range rebalancerStatus.MoveJobs {
		_, statuses[id] = cache.Target.GetJob(state.JobID(rebalancerStatus.MoveJobs[id]))
	}

	if err := r.actionCtx.WithStatusUpdate(ctx, func(s *api.DeploymentStatus) bool {
		s.Rebalancer.LastCheckTime = k8sutil.NewTime(meta.Now())

		var m []string

		for id := range rebalancerStatus.MoveJobs {
			if statuses[id] == state.JobPhaseFailed || statuses[id] == state.JobPhaseUnknown {
				r.log.Warn("Error while moving job")
				r.actionCtx.Metrics().GetRebalancer().AddFailures(1)
				continue
			}

			if statuses[id] == state.JobPhaseFinished {
				r.actionCtx.Metrics().GetRebalancer().AddSuccesses(1)
				continue
			}

			m = append(m, rebalancerStatus.MoveJobs[id])
		}

		s.Rebalancer.MoveJobs = m

		if len(s.Rebalancer.MoveJobs) == 0 {
			s.Rebalancer.LastCheckTime = nil
		}

		return true
	}); err != nil {
		r.log.Err(err).Warn("Unable to update status")
		return true, nil
	}
	return true, nil
}
