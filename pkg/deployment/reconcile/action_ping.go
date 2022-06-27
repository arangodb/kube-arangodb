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
	"time"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
)

func init() {
	registerAction(api.ActionTypePing, newPing, pingTimeout)
}

func newPing(action api.Action, actionCtx ActionContext) Action {
	a := &pingAction{}

	a.actionImpl = newActionImplDefRef(action, actionCtx)

	return a
}

type pingAction struct {
	actionImpl
}

func (a *pingAction) Start(ctx context.Context) (bool, error) {
	if a.action.TaskID == "" {
		a.log.Error("taskName parameter is missing")
		return true, nil
	}

	return false, nil
}

func (a *pingAction) CheckProgress(ctx context.Context) (bool, bool, error) {
	if a.action.TaskID == "" {
		a.log.Error("taskName parameter is missing")
		return false, true, nil
	}

	tasksCache, err := a.actionCtx.ACS().Cache().ArangoTask().V1()
	if err != nil {
		a.log.Err(err).Error("Failed to get ArangoTask cache")
		return false, false, err
	}

	task, exist := tasksCache.GetSimpleByID(a.action.TaskID)
	if !exist {
		a.log.Error("ArangoTask not found")
		return false, false, err
	}

	if task.Spec.Details != nil {
		pingBody := api.ArangoTaskPing{}
		if err := task.Spec.Details.Get(&pingBody); err != nil {
			a.log.Err(err).Error("Failed to parse ArangoTaskPing content")
			return false, false, err
		}

		if pingBody.DurationSeconds != 0 {
			a.log.Info("Checking ArangoTaskPing duration limits")
			upTo := a.action.CreationTime.Add(time.Duration(pingBody.DurationSeconds) * time.Second)
			if time.Now().Before(upTo) {
				return false, false, nil
			}
		}
	}

	return true, false, nil
}
