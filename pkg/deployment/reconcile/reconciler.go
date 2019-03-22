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

package reconcile

import (
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/rs/zerolog"
)

// Reconciler is the service that takes care of bring the a deployment
// in line with its (changed) specification.
type Reconciler struct {
	log     zerolog.Logger
	context Context
}

// NewReconciler creates a new reconciler with given context.
func NewReconciler(log zerolog.Logger, context Context) *Reconciler {
	return &Reconciler{
		log:     log,
		context: context,
	}
}

// CheckDeployment checks for obviously broken things and fixes them immediately
func (r *Reconciler) CheckDeployment() error {
	spec := r.context.GetSpec()
	status, _ := r.context.GetStatus()

	if spec.GetMode().HasCoordinators() {
		// Check if there are coordinators
		if len(status.Members.Coordinators) == 0 {
			// No more coordinators! Take immediate action
			r.log.Error().Msg("No Coordinator members! Create one member immediately")
			_, err := r.context.CreateMember(api.ServerGroupCoordinators, "")
			if err != nil {
				return err
			}
		} else if status.Members.Coordinators.AllFailed() {
			r.log.Error().Msg("All coordinators failed - reset")
			for _, m := range status.Members.Coordinators {
				if err := r.context.DeletePod(m.PodName); err != nil {
					r.log.Error().Err(err).Msg("Failed to delete pod")
				}
				m.Phase = api.MemberPhaseNone
				if err := status.Members.Update(m, api.ServerGroupCoordinators); err != nil {
					r.log.Error().Err(err).Msg("Failed to update member")
				}
			}
		}
	}

	return nil
}
