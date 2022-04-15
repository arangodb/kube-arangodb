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
	"github.com/rs/zerolog"
)

func init() {
	registerAction(api.ActionTypeSetMaintenanceCondition, newSetMaintenanceConditionAction, addMemberTimeout)
}

func newSetMaintenanceConditionAction(log zerolog.Logger, action api.Action, actionCtx ActionContext) Action {
	a := &actionSetMaintenanceCondition{}

	a.actionImpl = newActionImpl(log, action, actionCtx, &a.newMemberID)

	return a
}

type actionSetMaintenanceCondition struct {
	// actionImpl implement timeout and member id functions
	actionImpl

	actionEmptyCheckProgress

	newMemberID string
}

func (a *actionSetMaintenanceCondition) Start(ctx context.Context) (bool, error) {
	switch a.actionCtx.GetMode() {
	case api.DeploymentModeSingle:
		return true, nil
	}

	agencyState, agencyOK := a.actionCtx.GetAgencyCache()
	if !agencyOK {
		a.log.Error().Msgf("Unable to determine maintenance condition")
	} else {

		if err := a.actionCtx.WithStatusUpdate(ctx, func(s *api.DeploymentStatus) bool {
			if agencyState.Supervision.Maintenance.Exists() {
				return s.Conditions.Update(api.ConditionTypeMaintenanceMode, true, "Maintenance", "Maintenance enabled")
			} else {
				return s.Conditions.Remove(api.ConditionTypeMaintenanceMode)
			}
		}); err != nil {
			a.log.Error().Err(err).Msgf("Unable to set maintenance condition")
			return true, nil
		}
	}

	return true, nil
}
