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
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/k8s-operator/pkg/apis/arangodb/v1alpha"
	"github.com/arangodb/k8s-operator/pkg/util/k8sutil"
)

// executePlan tries to execute the plan as far as possible.
// Returns true when it has to be called again soon.
// False otherwise.
func (d *Deployment) executePlan(ctx context.Context) (bool, error) {
	log := d.deps.Log

	for {
		if len(d.status.Plan) == 0 {
			// No plan exists, nothing to be done
			return false, nil
		}

		// Take first action
		action := d.status.Plan[0]
		if action.StartTime.IsZero() {
			// Not started yet
			ready, err := d.startAction(ctx, action)
			if err != nil {
				log.Debug().Err(err).
					Str("action-type", string(action.Type)).
					Msg("Failed to start action")
				return false, maskAny(err)
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
				return false, maskAny(err)
			}
			if !ready {
				// We need to check back soon
				return true, nil
			}
			// Continue with next action
		} else {
			// First action of plan has been started, check its progress
			ready, err := d.checkActionProgress(ctx, action)
			if err != nil {
				log.Debug().Err(err).
					Str("action-type", string(action.Type)).
					Msg("Failed to check action progress")
				return false, maskAny(err)
			}
			if !ready {
				// Not ready check, come back soon
				return true, nil
			}
			// Remove action from list
			d.status.Plan = d.status.Plan[1:]
			// Save plan update
			if err := d.updateCRStatus(); err != nil {
				return false, maskAny(err)
			}
			// Continue with next action
		}
	}
}

// startAction performs the start of the given action
// Returns true if the action is completely finished, false in case
// the start time needs to be recorded and a ready condition needs to be checked.
func (d *Deployment) startAction(ctx context.Context, action api.Action) (bool, error) {
	log := d.deps.Log
	ns := d.apiObject.GetNamespace()

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
		m, _, ok := d.status.Members.ElementByID(action.MemberID)
		if !ok {
			// We wanted to remove and it is already gone. All ok
			return true, nil
		}
		// Remove the pod (if any)
		if err := d.deps.KubeCli.Core().Pods(ns).Delete(m.PodName, &metav1.DeleteOptions{}); err != nil && !k8sutil.IsNotFound(err) {
			log.Debug().Err(err).Str("pod", m.PodName).Msg("Failed to remove pod")
			return false, maskAny(err)
		}
		// Remove the pvc (if any)
		if m.PersistentVolumeClaimName != "" {
			if err := d.deps.KubeCli.Core().PersistentVolumeClaims(ns).Delete(m.PersistentVolumeClaimName, &metav1.DeleteOptions{}); err != nil && !k8sutil.IsNotFound(err) {
				log.Debug().Err(err).Str("pod", m.PodName).Msg("Failed to remove pvc")
				return false, maskAny(err)
			}
		}
		// Remove member
		if err := d.status.Members.RemoveByID(action.MemberID, action.Group); err != nil {
			log.Debug().Err(err).Str("group", action.Group.AsRole()).Msg("Failed to remove member")
			return false, maskAny(err)
		}
		// Save removed member
		if err := d.updateCRStatus(); err != nil {
			return false, maskAny(err)
		}
		return true, nil
	case api.ActionTypeCleanOutMember:
		m, ok := d.status.Members.DBServers.ElementByID(action.MemberID)
		if !ok {
			log.Error().Str("group", action.Group.AsRole()).Str("id", action.MemberID).Msg("No such member")
			return true, nil
		}
		c, err := d.clientCache.GetDatabase()
		if err != nil {
			log.Debug().Err(err).Str("group", action.Group.AsRole()).Msg("Failed to create member client")
			return false, maskAny(err)
		}
		cluster, err := c.Cluster(ctx)
		if err != nil {
			log.Debug().Err(err).Str("group", action.Group.AsRole()).Msg("Failed to access cluster")
			return false, maskAny(err)
		}
		if err := cluster.CleanOutServer(ctx, action.MemberID); err != nil {
			log.Debug().Err(err).Str("group", action.Group.AsRole()).Msg("Failed to cleanout member")
			return false, maskAny(err)
		}
		// Update status
		m.State = api.MemberStateCleanOut
		if err := d.updateCRStatus(); err != nil {
			return false, maskAny(err)
		}
		return true, nil
	case api.ActionTypeShutdownMember:
		m, _, ok := d.status.Members.ElementByID(action.MemberID)
		if !ok {
			log.Error().Str("group", action.Group.AsRole()).Str("id", action.MemberID).Msg("No such member")
			return true, nil
		}
		c, err := d.clientCache.Get(action.Group, action.MemberID)
		if err != nil {
			log.Debug().Err(err).Str("group", action.Group.AsRole()).Msg("Failed to create member client")
			return false, maskAny(err)
		}
		if err := c.Shutdown(ctx, true); err != nil {
			log.Debug().Err(err).Str("group", action.Group.AsRole()).Msg("Failed to shutdown member")
			return false, maskAny(err)
		}
		// Update status
		m.State = api.MemberStateShuttingDown
		if err := d.updateCRStatus(); err != nil {
			return false, maskAny(err)
		}
		return true, nil
	default:
		return false, maskAny(fmt.Errorf("Unknown action type"))
	}
}

// checkActionProgress checks the progress of the given action.
// Returns true if the action is completely finished, false otherwise.
func (d *Deployment) checkActionProgress(ctx context.Context, action api.Action) (bool, error) {
	switch action.Type {
	case api.ActionTypeAddMember:
		// Nothing todo
		return true, nil
	case api.ActionTypeRemoveMember:
		// Nothing todo
		return true, nil
	case api.ActionTypeCleanOutMember:
		c, err := d.clientCache.GetDatabase()
		if err != nil {
			return false, maskAny(err)
		}
		cluster, err := c.Cluster(ctx)
		if err != nil {
			return false, maskAny(err)
		}
		cleanedOut, err := cluster.IsCleanedOut(ctx, action.MemberID)
		if err != nil {
			return false, maskAny(err)
		}
		if !cleanedOut {
			// We're not done yet
			return false, nil
		}
		// Cleanout completed
		return true, nil
	case api.ActionTypeShutdownMember:
		m, _, ok := d.status.Members.ElementByID(action.MemberID)
		if !ok {
			// Member not long exists
			return true, nil
		}
		if m.Conditions.IsTrue(api.ConditionTypeTerminated) {
			// Shutdown completed
			return true, nil
		}
		// Member still not shutdown, retry soon
		return false, nil
	default:
		return false, maskAny(fmt.Errorf("Unknown action type"))
	}
}
