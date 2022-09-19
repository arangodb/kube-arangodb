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
)

func newEnableMaintenanceAction(action api.Action, actionCtx ActionContext) Action {
	a := &actionEnableMaintenance{}

	a.actionImpl = newActionImplDefRef(action, actionCtx)

	return a
}

type actionEnableMaintenance struct {
	// actionImpl implement timeout and member id functions
	actionImpl

	actionEmptyCheckProgress
}

func (a *actionEnableMaintenance) Start(ctx context.Context) (bool, error) {
	switch a.actionCtx.GetMode() {
	case api.DeploymentModeSingle:
		return true, nil
	}

	if err := a.actionCtx.SetAgencyMaintenanceMode(ctx, true); err != nil {
		a.log.Err(err).Error("Unable to enable maintenance")
		return true, nil
	}

	return true, nil
}
