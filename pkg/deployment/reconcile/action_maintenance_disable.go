//
// DISCLAIMER
//
// Copyright 2020-2021 ArangoDB GmbH, Cologne, Germany
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
// Author Adam Janikowski
// Author Tomasz Mielech
//

package reconcile

import (
	"context"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/rs/zerolog"
)

func init() {
	registerAction(api.ActionTypeDisableMaintenance, newDisableMaintenanceAction)
}

func newDisableMaintenanceAction(log zerolog.Logger, action api.Action, actionCtx ActionContext) Action {
	a := &actionDisableMaintenance{}

	a.actionImpl = newActionImpl(log, action, actionCtx, addMemberTimeout, &a.newMemberID)

	return a
}

type actionDisableMaintenance struct {
	// actionImpl implement timeout and member id functions
	actionImpl

	actionEmptyCheckProgress

	newMemberID string
}

func (a *actionDisableMaintenance) Start(ctx context.Context) (bool, error) {
	switch a.actionCtx.GetMode() {
	case api.DeploymentModeSingle:
		return true, nil
	}

	if err := a.actionCtx.SetAgencyMaintenanceMode(ctx, false); err != nil {
		a.log.Error().Err(err).Msgf("Unable to disable maintenance")
		return true, nil
	}

	return true, nil
}
