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
	"strconv"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
)

func newSetConditionAction(action api.Action, actionCtx ActionContext) Action {
	a := &actionSetCondition{}

	a.actionImpl = newActionImplDefRef(action, actionCtx)

	return a
}

type actionSetCondition struct {
	// actionImpl implement timeout and member id functions
	actionImpl

	actionEmptyCheckProgress
}

// Start starts the action for changing conditions on the provided member.
func (a actionSetCondition) Start(ctx context.Context) (bool, error) {
	if len(a.action.Params) == 0 {
		a.log.Info("can not start the action with the empty list of conditions")
		return true, nil
	}

	if err := a.actionCtx.WithStatusUpdate(ctx, func(s *api.DeploymentStatus) bool {
		changed := false
		for condition, value := range a.action.Params {
			if value == "" {
				a.log.Debug("remove the condition")

				if s.Conditions.Remove(api.ConditionType(condition)) {
					changed = true
				}
			} else {
				set, err := strconv.ParseBool(value)
				if err != nil {
					a.log.Err(err).Str("value", value).Error("can not parse string to boolean")
					continue
				}

				a.log.Debug("set the condition")

				if s.Conditions.Update(api.ConditionType(condition), set, a.action.Reason, "action set the member condition") {
					changed = true
				}
			}
		}
		return changed
	}); err != nil {
		a.log.Err(err).Warn("Unable to set condition")
		return true, nil
	}

	return true, nil
}
