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
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/strings"
)

type planGenerationOutput struct {
	plan    api.Plan
	backoff api.BackOff
	changed bool
	planner planner
}

type planGeneratorFunc func(ctx context.Context, apiObject k8sutil.APIObject,
	currentPlan api.Plan, spec api.DeploymentSpec,
	status api.DeploymentStatus,
	builderCtx PlanBuilderContext) (api.Plan, api.BackOff, bool)

type planGenerator func(ctx context.Context) planGenerationOutput

func (d *Reconciler) generatePlanFunc(gen planGeneratorFunc, planner planner) planGenerator {
	return func(ctx context.Context) planGenerationOutput {
		// Create plan
		apiObject := d.context.GetAPIObject()
		spec := d.context.GetSpec()
		status := d.context.GetStatus()
		builderCtx := newPlanBuilderContext(d.context)
		newPlan, backoffs, changed := gen(ctx, apiObject, planner.Get(&status), spec, status, builderCtx)

		return planGenerationOutput{
			plan:    newPlan,
			backoff: backoffs,
			planner: planner,
			changed: changed,
		}
	}
}

func (d *Reconciler) generatePlan(ctx context.Context, generators ...planGenerator) (error, bool) {
	updated := false
	updateRequired := false

	if err := d.context.WithStatusUpdate(ctx, func(s *api.DeploymentStatus) bool {
		var b api.BackOff

		for id := range generators {
			result := generators[id](ctx)

			b = b.CombineLatest(result.backoff)

			if len(result.plan) == 0 || !result.changed {
				continue
			}

			// Send events
			current := result.planner.Get(s)

			for id := len(current); id < len(result.plan); id++ {
				action := result.plan[id]
				d.context.CreateEvent(k8sutil.NewPlanAppendEvent(d.context.GetAPIObject(), action.Type.String(), action.Group.AsRole(), action.MemberID, action.Reason))
				if r := action.Reason; r != "" {
					d.log.Str("Action", action.Type.String()).
						Str("Role", action.Group.AsRole()).Str("Member", action.MemberID).
						Str("Type", strings.Title(result.planner.Type())).Info(r)
				}
			}

			result.planner.Set(s, result.plan)

			for _, p := range result.plan {
				actionsGeneratedMetrics.WithLabelValues(d.context.GetName(), p.Type.String(), result.planner.Type()).Inc()
			}

			updated = true
		}

		if len(b) > 0 {
			new := s.BackOff.DeepCopy().Combine(b)

			if !new.Equal(s.BackOff) {
				s.BackOff = new
				updateRequired = true
			}
		}

		return updated || updateRequired
	}); err != nil {
		return errors.WithMessage(err, "Unable to save plan"), false
	}

	return nil, updated
}
