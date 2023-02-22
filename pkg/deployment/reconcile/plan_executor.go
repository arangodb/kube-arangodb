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
	"fmt"
	"time"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/metrics"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/errors/panics"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
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

type plannerResources struct {
}

func (p plannerResources) Type() string {
	return "resources"
}

func (p plannerResources) Get(deployment *api.DeploymentStatus) api.Plan {
	return deployment.ResourcesPlan
}

func (p plannerResources) Set(deployment *api.DeploymentStatus, plan api.Plan) bool {
	if !deployment.ResourcesPlan.Equal(plan) {
		deployment.ResourcesPlan = plan
		return true
	}

	return false
}

// ExecutePlan tries to execute the plan as far as possible.
// Returns true when it has to be called again soon.
// False otherwise.
func (d *Reconciler) ExecutePlan(ctx context.Context) (bool, error) {
	execution := 0

	var retrySoon bool

	for {
		if execution >= 32 {
			return retrySoon, nil
		}

		execution++

		if retrySoonCall, recall, err := d.executePlanInLoop(ctx); err != nil {
			return false, err
		} else {
			retrySoon = retrySoon || retrySoonCall
			if recall {
				continue
			}
		}

		break
	}

	return retrySoon, nil
}

// ExecutePlan tries to execute the plan as far as possible.
// Returns true when it has to be called again soon.
// False otherwise.
func (d *Reconciler) executePlanInLoop(ctx context.Context) (bool, bool, error) {
	var callAgain bool
	var callInLoop bool

	if again, inLoop, err := d.executePlanStatus(ctx, plannerHigh{}); err != nil {
		d.planLogger.Err(err).Error("Execution of plan failed")
		return false, false, errors.WithStack(err)
	} else {
		callAgain = callAgain || again
		callInLoop = callInLoop || inLoop
	}

	if again, inLoop, err := d.executePlanStatus(ctx, plannerResources{}); err != nil {
		d.planLogger.Err(err).Error("Execution of plan failed")
		return false, false, nil
	} else {
		callAgain = callAgain || again
		callInLoop = callInLoop || inLoop
	}

	if again, inLoop, err := d.executePlanStatus(ctx, plannerNormal{}); err != nil {
		d.planLogger.Err(err).Error("Execution of plan failed")
		return false, false, errors.WithStack(err)
	} else {
		callAgain = callAgain || again
		callInLoop = callInLoop || inLoop
	}

	return callAgain, callInLoop, nil
}

func (d *Reconciler) executePlanStatus(ctx context.Context, pg planner) (bool, bool, error) {
	loopStatus := d.context.GetStatus()

	plan := pg.Get(&loopStatus)

	if len(plan) == 0 {
		return false, false, nil
	}

	newPlan, callAgain, callInLoop, err := d.executePlan(ctx, plan, pg)

	// Refresh current status
	loopStatus = d.context.GetStatus()

	if pg.Set(&loopStatus, newPlan) {
		d.planLogger.Info("Updating plan")
		if err := d.context.UpdateStatus(ctx, loopStatus); err != nil {
			d.planLogger.Err(err).Debug("Failed to update CR status")
			return false, false, errors.WithStack(err)
		}
	}

	if err != nil {
		return false, false, err
	}

	return callAgain, callInLoop, nil
}

func (d *Reconciler) executePlan(ctx context.Context, statusPlan api.Plan, pg planner) (newPlan api.Plan, callAgain, callInLoop bool, err error) {
	plan := statusPlan.DeepCopy()

	for {
		if len(plan) == 0 {
			return nil, false, false, nil
		}

		// Take first action
		planAction := plan[0]

		action, actionContext := d.createAction(planAction)

		if !planAction.IsStarted() {
			if err := getActionPre(action, ctx); err != nil {
				d.planLogger.Err(err).Error("Pre action failed")
				return nil, false, false, errors.WithStack(err)
			}
		}

		done, abort, recall, retry, err := d.executeAnyAction(ctx, planAction, action)
		if err != nil {
			if retry {
				return plan, true, false, nil
			}
			// The Plan will be cleaned up, so no actions will be in the queue.
			actionsCurrentPlan.WithLabelValues(d.context.GetName(), planAction.Group.AsRole(), planAction.MemberID,
				planAction.Type.String(), pg.Type()).Set(0.0)

			actionsFailedMetrics.WithLabelValues(d.context.GetName(), planAction.Type.String(), pg.Type()).Inc()
			return nil, false, false, errors.WithStack(err)
		}

		if abort {
			// The Plan will be cleaned up, so no actions will be in the queue.
			actionsCurrentPlan.WithLabelValues(d.context.GetName(), planAction.Group.AsRole(), planAction.MemberID,
				planAction.Type.String(), pg.Type()).Set(0.0)

			actionsFailedMetrics.WithLabelValues(d.context.GetName(), planAction.Type.String(), pg.Type()).Inc()
			return nil, true, false, nil
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

			if uid, components := getActionReloadCachedStatus(action); len(components) > 0 {
				c, ok := d.context.ACS().ClusterCache(uid)
				if ok {
					c.GetThrottles().Invalidate(components...)

					d.planLogger.Info("Reloading cached status")
					if err := c.Refresh(ctx); err != nil {
						d.planLogger.Err(err).Warn("Unable to reload cached status")
						return plan, recall, false, nil
					}
				}
			}

			if newPlan, changed := getActionPlanAppender(action, plan); changed {
				// Our actions have been added to the end of plan
				return newPlan, false, true, nil
			}

			if err := getActionPost(action, ctx); err != nil {
				d.planLogger.Err(err).Error("Post action failed")
				return nil, false, false, errors.WithStack(err)
			}
		} else {
			if !plan[0].IsStarted() {
				// The action has been started in this iteration, but it is not finished yet.
				actionsCurrentPlan.WithLabelValues(d.context.GetName(), planAction.Group.AsRole(), planAction.MemberID,
					planAction.Type.String(), pg.Type()).Inc()

				now := meta.Now()
				plan[0].StartTime = &now
			}

			plan[0].Locals.Merge(actionContext.CurrentLocals())
			plan[0].Progress = actionContext.GetProgress()
			return plan, recall, false, nil
		}
	}
}

func (d *Reconciler) executeAnyAction(ctx context.Context, planAction api.Action, action Action) (done, abort, callAgain, retry bool, err error) {
	if planAction.Type.Optional() {
		return d.executeOptionalAction(ctx, planAction, action)
	}

	return d.executeAction(ctx, planAction, action)
}

func (d *Reconciler) executeOptionalAction(ctx context.Context, planAction api.Action, action Action) (done, abort, callAgain, retry bool, err error) {
	done, abort, callAgain, retry, err = d.executeAction(ctx, planAction, action)
	if err != nil {
		// Check if we still can retry
		if retry {
			return
		}

		// We cant retry anymore, check if abort was requested
		if abort {
			return
		}

		done = true
		err = nil
		d.planLogger.Str("action", string(planAction.Type)).Str("member", planAction.MemberID).Err(err).Warn("Optional action failed, but it is safe to continue")
	}

	return
}

func (d *Reconciler) executeAction(ctx context.Context, planAction api.Action, action Action) (done, abort, callAgain, retry bool, err error) {
	log := d.planLogger.Str("action", string(planAction.Type)).Str("member", planAction.MemberID)
	log.Info("Executing action")

	if !planAction.IsStarted() {
		// Not started yet
		ready, err := d.executeActionStart(ctx, action)
		if err != nil {
			if g := getStartFailureGracePeriod(action); g > 0 && !planAction.CreationTime.IsZero() {
				if time.Since(planAction.CreationTime.Time) < g {
					log.Err(err).Error("Failed to start action, but still in grace period")
					return false, false, false, true, errors.WithStack(err)
				}
			}

			log.Err(err).Error("Failed to start action")
			return false, false, false, false, errors.WithStack(err)
		}

		if ready {
			log.Bool("ready", ready).Info("Action Start completed")
			return true, false, false, false, nil
		}
		log.Bool("ready", ready).Info("Action Started")

		return false, false, true, false, nil
	}

	timeout := GetActionTimeout(d.context.GetSpec(), planAction.Type)

	if t := planAction.StartTime; t != nil {
		if tm := t.Time; !tm.IsZero() {
			since := time.Since(tm)
			log = log.Dur("duration", since)
			if !timeout.Infinite() {
				log = log.Dur("timeouts_in", timeout.Duration-since)
			}
		}
	}

	// First action of plan has been started, check its progress
	ready, abort, err := d.executeActionCheckProgress(ctx, action)
	if err != nil {
		log.Err(err).Debug("Failed to check action progress")
		return false, false, false, false, errors.WithStack(err)
	}

	log.
		Bool("abort", abort).
		Bool("ready", ready).
		Debug("Action CheckProgress completed")

	if ready {
		return true, false, false, false, nil
	}

	if abort {
		log.Warn("Action aborted. Removing the entire plan")
		d.context.CreateEvent(k8sutil.NewPlanAbortedEvent(d.context.GetAPIObject(), string(planAction.Type), planAction.MemberID, planAction.Group.AsRole()))
		return false, true, false, false, nil
	} else if isActionTimeout(timeout, planAction) {
		log.Warn("Action not finished in time. Removing the entire plan")
		d.context.CreateEvent(k8sutil.NewPlanTimeoutEvent(d.context.GetAPIObject(), string(planAction.Type), planAction.MemberID, planAction.Group.AsRole()))
		return false, true, false, false, nil

	}

	// Timeout not yet expired, come back soon
	return false, false, true, false, nil
}

func isActionTimeout(timeout api.Timeout, planAction api.Action) bool {
	if planAction.StartTime == nil {
		return false
	}
	if planAction.StartTime.IsZero() {
		return false
	}
	if timeout.Infinite() {
		return false
	}
	return time.Since(planAction.StartTime.Time) > timeout.Duration
}

func (d *Reconciler) executeActionCheckProgress(ctx context.Context, action Action) (ready bool, abort bool, retErr error) {
	retErr = panics.RecoverWithSection("ActionProgress", func() (err error) {
		ready, abort, err = action.CheckProgress(ctx)
		return
	})

	return
}

func (d *Reconciler) executeActionStart(ctx context.Context, action Action) (done bool, retErr error) {
	retErr = panics.RecoverWithSection("ActionStart", func() (err error) {
		done, err = action.Start(ctx)
		return
	})

	return
}

// createAction create action object based on action type
func (d *Reconciler) createAction(action api.Action) (Action, ActionContext) {
	actionCtx := newActionContext(d.log, d.context, &d.metrics)

	f, ok := getActionFactory(action.Type)
	if !ok {
		panic(fmt.Sprintf("Unknown action type '%s'", action.Type))
	}

	return f(action, actionCtx), actionCtx
}
