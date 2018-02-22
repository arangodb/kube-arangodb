//
// DISCLAIMER
//
// Copyright 2018 ArangoDB GmbH, Cologne, Germany
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

package deployment

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/k8s-operator/pkg/apis/arangodb/v1alpha"
)

// executePlan tries to execute the plan as far as possible.
func (d *Deployment) executePlan() error {
	log := d.deps.Log

	for {
		if len(d.status.Plan) == 0 {
			// No plan exists, nothing to be done
			return nil
		}

		// Take first action
		action := d.status.Plan[0]
		if action.StartTime.IsZero() {
			// Not started yey
			ready, err := d.startAction(action)
			if err != nil {
				log.Debug().Err(err).
					Str("action-type", string(action.Type)).
					Msg("Failed to start action")
				return maskAny(err)
			}
			if ready {
				// Remove action from list
				d.status.Plan = d.status.Plan[1:]
			} else {
				// Mark start time
				now := metav1.Now()
				d.status.Plan[0].StartTime = &now
			}
			// Save plan update
			if err := d.updateCRStatus(); err != nil {
				return maskAny(err)
			}
		} else {
			// First action of plan has been started, check its progress
			ready, err := d.checkActionProgress(action)
			if err != nil {
				log.Debug().Err(err).
					Str("action-type", string(action.Type)).
					Msg("Failed to check action progress")
				return maskAny(err)
			}
			if ready {
				// Remove action from list
				d.status.Plan = d.status.Plan[1:]
				// Save plan update
				if err := d.updateCRStatus(); err != nil {
					return maskAny(err)
				}
			}
		}
	}
}

// startAction performs the start of the given action
// Returns true if the action is completely finished, false in case
// the start time needs to be recorded and a ready condition needs to be checked.
func (d *Deployment) startAction(action api.Action) (bool, error) {
	log := d.deps.Log

	switch action.Type {
	case api.ActionTypeAddMember:
		if err := d.createMember(action.Group, d.apiObject); err != nil {
			log.Debug().Err(err).Str("group", action.Group.AsRole()).Msg("Failed to create member")
			return false, maskAny(err)
		}
		// Save added member
		if err := d.updateCRStatus(); err != nil {
			return false, maskAny(err)
		}
		return true, nil
	case api.ActionTypeRemoveMember:
		if err := d.status.Members.RemoveByID(action.MemberID, action.Group); api.IsNotFound(err) {
			// We wanted to remove and it is already gone. All ok
			return true, nil
		} else if err != nil {
			log.Debug().Err(err).Str("group", action.Group.AsRole()).Msg("Failed to remove member")
			return false, maskAny(err)
		}
		// Save removed member
		if err := d.updateCRStatus(); err != nil {
			return false, maskAny(err)
		}
		return true, nil
	case api.ActionTypeDrainMember:
		// TODO
		return true, nil
	case api.ActionTypeShutdownMember:
		// TODO
		return true, nil
	default:
		return false, maskAny(fmt.Errorf("Unknown action type"))
	}
}

// checkActionProgress checks the progress of the given action.
// Returns true if the action is completely finished, false otherwise.
func (d *Deployment) checkActionProgress(action api.Action) (bool, error) {
	switch action.Type {
	case api.ActionTypeAddMember:
		// Nothing todo
		return true, nil
	case api.ActionTypeRemoveMember:
		// Nothing todo
		return true, nil
	case api.ActionTypeDrainMember:
		// TODO
		return true, nil
	case api.ActionTypeShutdownMember:
		// TODO
		return true, nil
	default:
		return false, maskAny(fmt.Errorf("Unknown action type"))
	}
}
