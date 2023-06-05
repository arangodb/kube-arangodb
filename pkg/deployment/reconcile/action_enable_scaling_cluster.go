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

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

// newEnableClusterScalingAction creates the new action with enabling scaling DBservers and coordinators.
func newEnableClusterScalingAction(action api.Action, actionCtx ActionContext) Action {
	a := &actionEnableClusterScaling{}

	a.actionImpl = newActionImpl(action, actionCtx, util.NewType[string](""))

	return a
}

// actionEnableClusterScaling implements enabling scaling DBservers and coordinators.
type actionEnableClusterScaling struct {
	// actionImpl implement timeout and member id functions
	actionImpl

	// actionEmptyCheckProgress implement check progress with empty implementation
	actionEmptyCheckProgress
}

// Start enables scaling DBservers and coordinators
func (a *actionEnableClusterScaling) Start(ctx context.Context) (bool, error) {
	err := a.actionCtx.EnableScalingCluster(ctx)
	if err != nil {
		return false, err
	}
	return true, nil
}
