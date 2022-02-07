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
	"fmt"
	"time"

	"github.com/rs/zerolog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/metrics"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
)

var (
	actionsGeneratedMetrics = metrics.MustRegisterCounterVec(reconciliationComponent, "actions_generated", "Number of actions added to the plan", metrics.DeploymentName, metrics.ActionName, metrics.ActionPriority)
	actionsSucceededMetrics = metrics.MustRegisterCounterVec(reconciliationComponent, "actions_succeeded", "Number of succeeded actions", metrics.DeploymentName, metrics.ActionName, metrics.ActionPriority)
	actionsFailedMetrics    = metrics.MustRegisterCounterVec(reconciliationComponent, "actions_failed", "Number of failed actions", metrics.DeploymentName, metrics.ActionName, metrics.ActionPriority)
	actionsCurrentPlan      = metrics.MustRegisterGaugeVec(reconciliationComponent, "actions_current",
		"The current number of the plan actions are being performed",
		metrics.DeploymentName, "group", "member", "name", "priority")
)

type planner interface {
	Get(deployment *api.DeploymentStatus) api.Plan
	Set(deployment *api.DeploymentStatus, p api.Plan) bool

	Type() string
}

var _ planner = plannerNormal{}
var _ planner = plannerHigh{}

type plannerNormal struct {
}

func (p plannerNormal) Type() string {
	return "normal"
}

func (p plannerNormal) Get(deployment *api.DeploymentStatus) api.Plan {
	return deployment.Plan
}

func (p plannerNormal) Set(deployment *api.DeploymentStatus, plan api.Plan) bool {
	if !deployment.Plan.Equal(plan) {
		deployment.Plan = plan
		return true
	}

	return false
}

type plannerHigh struct {
}

func (p plannerHigh) Type() string {
	return "high"
}

func (p plannerHigh) Get(deployment *api.DeploymentStatus) api.Plan {
	return deployment.HighPriorityPlan
}

func (p plannerHigh) Set(deployment *api.DeploymentStatus, plan api.Plan) bool {
	if !deployment.HighPriorityPlan.Equal(plan) {
		deployment.HighPriorityPlan = plan
		return true
	}

	return false
}

// ExecutePlan tries to execute the plan as far as possible.
// Returns true when it has to be called again soon.
// False otherwise.
func (d *Reconciler) ExecutePlan(ctx context.Context, cachedStatus inspectorInterface.Inspector) (bool, error) {
	var callAgain bool

	if again, err := d.executePlanStatus(ctx, cachedStatus, d.log, plannerHigh{}); err != nil {
		return false, errors.WithStack(err)
	} else if again {
		callAgain = true
	}

	if again, err := d.executePlanStatus(ctx, cachedStatus, d.log, plannerNormal{}); err != nil {
		return false, errors.WithStack(err)
	} else if again {
		callAgain = true
	}

	return callAgain, nil
}

func (d *Reconciler) executePlanStatus(ctx context.Context, cachedStatus inspectorInterface.Inspector, log zerolog.Logger, pg planner) (bool, error) {
	loopStatus, _ := d.context.GetStatus()

	plan := pg.Get(&loopStatus)

	if len(plan) == 0 {
		return false, nil
	}

	newPlan, callAgain, err := d.executePlan(ctx, cachedStatus, log, plan, pg)

	// Refresh current status
	loopStatus, lastVersion := d.context.GetStatus()

	if pg.Set(&loopStatus, newPlan) {
		log.Info().Msg("Updating plan")
		if err := d.context.UpdateStatus(ctx, loopStatus, lastVersion, true); err != nil {
			log.Debug().Err(err).Msg("Failed to update CR status")
			return false, errors.WithStack(err)
		}
	}

	if err != nil {
		return false, err
	}

	return callAgain, nil
}

func (d *Reconciler) executePlan(ctx context.Context, cachedStatus inspectorInterface.Inspector, log zerolog.Logger, statusPlan api.Plan, pg planner) (newPlan api.Plan, callAgain bool, err error) {
	plan := statusPlan.DeepCopy()

	for {
		if len(plan) == 0 {
			return nil, false, nil
		}

		// Take first action
		planAction := plan[0]
		logContext := log.With().
			Int("plan-len", len(plan)).
			Str("action-id", planAction.ID).
			Str("action-type", string(planAction.Type)).
			Str("group", planAction.Group.AsRole()).
			Str("member-id", planAction.MemberID)

		if status, _ := d.context.GetStatus(); status.Members.ContainsID(planAction.MemberID) {
			if member, _, ok := status.Members.ElementByID(planAction.MemberID); ok {
				logContext = logContext.Str("phase", string(member.Phase))
			}
		}

		for k, v := range planAction.Params {
			logContext = logContext.Str(k, v)
		}

		log := logContext.Logger()

		action := d.createAction(log, planAction, cachedStatus)

		done, abort, recall, retry, err := d.executeAction(ctx, log, planAction, action)
		if err != nil {
			if retry {
				return plan, true, nil
			}
			// The Plan will be cleaned up, so no actions will be in the queue.
			actionsCurrentPlan.WithLabelValues(d.context.GetName(), planAction.Group.AsRole(), planAction.MemberID,
				planAction.Type.String(), pg.Type()).Set(0.0)

			actionsFailedMetrics.WithLabelValues(d.context.GetName(), planAction.Type.String(), pg.Type()).Inc()
			return nil, false, errors.WithStack(err)
		}

		if abort {
			// The Plan will be cleaned up, so no actions will be in the queue.
			actionsCurrentPlan.WithLabelValues(d.context.GetName(), planAction.Group.AsRole(), planAction.MemberID,
				planAction.Type.String(), pg.Type()).Set(0.0)

			actionsFailedMetrics.WithLabelValues(d.context.GetName(), planAction.Type.String(), pg.Type()).Inc()
			return nil, true, nil
		}

		if done {
			if planAction.IsStarted() {
				// The below metrics was increased in the previous iteration, so it should be decreased now.
				// If the action hasn't been started in this iteration then the metrics have not been increased.
				actionsCurrentPlan.WithLabelValues(d.context.GetName(), planAction.Group.AsRole(), planAction.MemberID,
					planAction.Type.String(), pg.Type()).Dec()
			}

			actionsSucceededMetrics.WithLabelValues(d.context.GetName(), planAction.Type.String(), pg.Type()).Inc()
			if len(plan) > 1 {
				plan = plan[1:]
				if plan[0].MemberID == api.MemberIDPreviousAction {
					plan[0].MemberID = action.MemberID()
				}
			} else {
				actionsCurrentPlan.WithLabelValues(d.context.GetName(), planAction.Group.AsRole(), planAction.MemberID,
					planAction.Type.String(), pg.Type()).Set(0.0)
				plan = nil
			}

			if getActionReloadCachedStatus(action) {
				log.Info().Msgf("Reloading cached status")
				if err := cachedStatus.Refresh(ctx); err != nil {
					log.Warn().Err(err).Msgf("Unable to reload cached status")
					return plan, recall, nil
				}
			}

			if newPlan, changed := getActionPlanAppender(action, plan); changed {
				// Our actions have been added to the end of plan
				log.Info().Msgf("Appending new plan items")
				return newPlan, true, nil
			}

			if err := getActionPost(action, ctx); err != nil {
				log.Err(err).Msgf("Post action failed")
				return nil, false, errors.WithStack(err)
			}
		} else {
			if !plan[0].IsStarted() {
				// The action has been started in this iteration, but it is not finished yet.
				actionsCurrentPlan.WithLabelValues(d.context.GetName(), planAction.Group.AsRole(), planAction.MemberID,
					planAction.Type.String(), pg.Type()).Inc()

				now := metav1.Now()
				plan[0].StartTime = &now
			}

			return plan, recall, nil
		}
	}
}

func (d *Reconciler) executeAction(ctx context.Context, log zerolog.Logger, planAction api.Action, action Action) (done, abort, callAgain, retry bool, err error) {
	if !planAction.IsStarted() {
		// Not started yet
		ready, err := action.Start(ctx)
		if err != nil {
			if d := getStartFailureGracePeriod(action); d > 0 && !planAction.CreationTime.IsZero() {
				if time.Since(planAction.CreationTime.Time) < d {
					log.Error().Err(err).Msg("Failed to start action, but still in grace period")
					return false, false, false, true, errors.WithStack(err)
				}
			}
			log.Error().Err(err).Msg("Failed to start action")
			return false, false, false, false, errors.WithStack(err)
		}

		if ready {
			log.Debug().Bool("ready", ready).Msg("Action Start completed")
			return true, false, false, false, nil
		}

		return false, false, true, false, nil
	}
	// First action of plan has been started, check its progress
	ready, abort, err := action.CheckProgress(ctx)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to check action progress")
		return false, false, false, false, errors.WithStack(err)
	}

	log.Debug().
		Bool("abort", abort).
		Bool("ready", ready).
		Msg("Action CheckProgress completed")

	if ready {
		return true, false, false, false, nil
	}

	if abort {
		log.Warn().Msg("Action aborted. Removing the entire plan")
		d.context.CreateEvent(k8sutil.NewPlanAbortedEvent(d.context.GetAPIObject(), string(planAction.Type), planAction.MemberID, planAction.Group.AsRole()))
		return false, true, false, false, nil
	} else if time.Now().After(planAction.CreationTime.Add(action.Timeout(d.context.GetSpec()))) {
		log.Warn().Msg("Action not finished in time. Removing the entire plan")
		d.context.CreateEvent(k8sutil.NewPlanTimeoutEvent(d.context.GetAPIObject(), string(planAction.Type), planAction.MemberID, planAction.Group.AsRole()))
		return false, true, false, false, nil
	}

	// Timeout not yet expired, come back soon
	return false, false, true, false, nil
}

// createAction create action object based on action type
func (d *Reconciler) createAction(log zerolog.Logger, action api.Action, cachedStatus inspectorInterface.Inspector) Action {
	actionCtx := newActionContext(log.With().Str("id", action.ID).Str("type", action.Type.String()).Logger(), d.context, cachedStatus)

	f, ok := getActionFactory(action.Type)
	if !ok {
		panic(fmt.Sprintf("Unknown action type '%s'", action.Type))
	}

	return f(log, action, actionCtx)
}
