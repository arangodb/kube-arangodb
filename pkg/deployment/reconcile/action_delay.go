//
// DISCLAIMER
//
// Copyright 2025 ArangoDB GmbH, Cologne, Germany
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
	"time"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
)

const (
	DelayActionDuration = "delayActionDuration"
)

// newDelayAction creates a new Action that implements the given
// planned Delay action.
func newDelayAction(action api.Action, actionCtx ActionContext) Action {
	a := &actionDelay{}

	a.actionImpl = newActionImplDefRef(action, actionCtx)

	return a
}

// actionDelay implements an Delay.
type actionDelay struct {
	// actionImpl implement timeout and member id functions
	actionImpl
}

// Start performs the start of the action.
// Returns true if the action is completely finished, false in case
// the start time needs to be recorded and a ready condition needs to be checked.
func (a *actionDelay) Start(ctx context.Context) (bool, error) {
	return false, nil
}

func (a *actionDelay) CheckProgress(ctx context.Context) (bool, bool, error) {
	v, ok := a.action.Params[DelayActionDuration]
	if !ok {
		a.log.Str("key", DelayActionDuration).Warn("Param for the delay not defined")
		return true, false, nil
	}

	d, err := time.ParseDuration(v)
	if err != nil {
		a.log.Err(err).Str("value", v).Warn("Unable to parse duration")
		return true, false, nil
	}

	if v := a.action.StartTime; v != nil {
		if v.Time.Add(d).Before(time.Now()) {
			return true, false, nil
		}
	}

	return false, false, nil
}
