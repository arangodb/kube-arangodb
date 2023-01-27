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

	core "k8s.io/api/core/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/reconcile/shared"
)

func newSetConditionV2Action(action api.Action, actionCtx ActionContext) Action {
	a := &actionSetConditionV2{}

	a.actionImpl = newActionImplDefRef(action, actionCtx)

	return a
}

type actionSetConditionV2 struct {
	// actionImpl implement timeout and member id functions
	actionImpl

	actionEmptyCheckProgress
}

// Start starts the action for changing conditions on the provided member.
func (a actionSetConditionV2) Start(ctx context.Context) (bool, error) {
	at, ok := a.action.Params[shared.SetConditionActionV2KeyType]
	if !ok {
		a.log.Info("key %s is missing in action definition", shared.SetConditionActionV2KeyType)
		return true, nil
	}

	aa, ok := a.action.Params[shared.SetConditionActionV2KeyAction]
	if !ok {
		a.log.Info("key %s is missing in action definition", shared.SetConditionActionV2KeyAction)
		return true, nil
	}

	switch at {
	case shared.SetConditionActionV2KeyTypeAdd:
		ah := a.action.Params[shared.SetConditionActionV2KeyHash]
		am := a.action.Params[shared.SetConditionActionV2KeyMessage]
		ar := a.action.Params[shared.SetConditionActionV2KeyReason]
		as := a.action.Params[shared.SetConditionActionV2KeyStatus] == string(core.ConditionTrue)

		if err := a.actionCtx.WithStatusUpdateErr(ctx, func(s *api.DeploymentStatus) (bool, error) {
			return s.Conditions.UpdateWithHash(api.ConditionType(aa), as, ar, am, ah), nil
		}); err != nil {
			a.log.Err(err).Warn("unable to update status")
			return true, nil
		}
	case shared.SetConditionActionV2KeyTypeRemove:
		if err := a.actionCtx.WithStatusUpdateErr(ctx, func(s *api.DeploymentStatus) (bool, error) {
			return s.Conditions.Remove(api.ConditionType(aa)), nil
		}); err != nil {
			a.log.Err(err).Warn("unable to update status")
			return true, nil
		}
	default:
		a.log.Info("unknown type %s", at)
		return true, nil
	}
	return true, nil
}
