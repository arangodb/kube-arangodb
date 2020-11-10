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
// Author Ewout Prangsma
//

package reconcile

import (
	"context"
	"fmt"
	"time"

	"github.com/arangodb/kube-arangodb/pkg/deployment/resources/inspector"

	"github.com/rs/zerolog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

// ExecutePlan tries to execute the plan as far as possible.
// Returns true when it has to be called again soon.
// False otherwise.
func (d *Reconciler) ExecutePlan(ctx context.Context, cachedStatus inspector.Inspector) (bool, error) {
	log := d.log
	firstLoop := true

	for {
		loopStatus, _ := d.context.GetStatus()
		if len(loopStatus.Plan) == 0 {
			// No plan exists or all action have finished, nothing to be done
			if !firstLoop {
				log.Debug().Msg("Reconciliation plan has finished")
			}
			return !firstLoop, nil
		}
		firstLoop = false

		// Take first action
		planAction := loopStatus.Plan[0]
		logContext := log.With().
			Int("plan-len", len(loopStatus.Plan)).
			Str("action-id", planAction.ID).
			Str("action-type", string(planAction.Type)).
			Str("group", planAction.Group.AsRole()).
			Str("member-id", planAction.MemberID)

		for k, v := range planAction.Params {
			logContext = logContext.Str(k, v)
		}

		log := logContext.Logger()

		action := d.createAction(ctx, log, planAction, cachedStatus)
		if planAction.StartTime.IsZero() {
			// Not started yet
			ready, err := action.Start(ctx)
			if err != nil {
				log.Debug().Err(err).
					Msg("Failed to start action")
				return false, maskAny(err)
			}
			{ // action.Start may have changed status, so reload it.
				status, lastVersion := d.context.GetStatus()
				// Update status according to result on action.Start.
				if ready {
					// Remove action from list
					status.Plan = status.Plan[1:]
					if len(status.Plan) > 0 && status.Plan[0].MemberID == api.MemberIDPreviousAction {
						// Fill in MemberID from previous action
						status.Plan[0].MemberID = action.MemberID()
					}
				} else {
					// Mark start time
					now := metav1.Now()
					status.Plan[0].StartTime = &now
				}
				// Save plan update
				if err := d.context.UpdateStatus(status, lastVersion, true); err != nil {
					log.Debug().Err(err).Msg("Failed to update CR status")
					return false, maskAny(err)
				}
			}
			log.Debug().Bool("ready", ready).Msg("Action Start completed")

			return true, nil
		} else {
			// First action of plan has been started, check its progress
			ready, abort, err := action.CheckProgress(ctx)
			if err != nil {
				log.Debug().Err(err).Msg("Failed to check action progress")
				return false, maskAny(err)
			}
			if ready {
				{ // action.CheckProgress may have changed status, so reload it.
					status, lastVersion := d.context.GetStatus()
					// Remove action from list
					status.Plan = status.Plan[1:]
					if len(status.Plan) > 0 && status.Plan[0].MemberID == api.MemberIDPreviousAction {
						// Fill in MemberID from previous action
						status.Plan[0].MemberID = action.MemberID()
					}
					// Save plan update
					if err := d.context.UpdateStatus(status, lastVersion); err != nil {
						log.Debug().Err(err).Msg("Failed to update CR status")
						return false, maskAny(err)
					}
				}
			}
			log.Debug().
				Bool("abort", abort).
				Bool("ready", ready).
				Msg("Action CheckProgress completed")
			if !ready {
				deadlineExpired := false
				if abort {
					log.Warn().Msg("Action aborted. Removing the entire plan")
					d.context.CreateEvent(k8sutil.NewPlanAbortedEvent(d.context.GetAPIObject(), string(planAction.Type), planAction.MemberID, planAction.Group.AsRole()))
				} else {
					// Not ready yet & no abort, check timeout
					deadline := planAction.CreationTime.Add(action.Timeout(d.context.GetSpec()))
					if time.Now().After(deadline) {
						// Timeout has expired
						deadlineExpired = true
						log.Warn().Msg("Action not finished in time. Removing the entire plan")
						d.context.CreateEvent(k8sutil.NewPlanTimeoutEvent(d.context.GetAPIObject(), string(planAction.Type), planAction.MemberID, planAction.Group.AsRole()))
					}
				}
				if abort || deadlineExpired {
					// Replace plan with empty one and save it.
					status, lastVersion := d.context.GetStatus()
					status.Plan = api.Plan{}
					if err := d.context.UpdateStatus(status, lastVersion); err != nil {
						log.Debug().Err(err).Msg("Failed to update CR status")
						return false, maskAny(err)
					}
					return true, nil
				}
				// Timeout not yet expired, come back soon
				return true, nil
			}
			return true, nil
		}
	}
}

// createAction create action object based on action type
func (d *Reconciler) createAction(ctx context.Context, log zerolog.Logger, action api.Action, cachedStatus inspector.Inspector) Action {
	actionCtx := newActionContext(log.With().Str("id", action.ID).Str("type", action.Type.String()).Logger(), d.context, cachedStatus)

	f, ok := getActionFactory(action.Type)
	if !ok {
		panic(fmt.Sprintf("Unknown action type '%s'", action.Type))
	}

	return f(log, action, actionCtx)
}
