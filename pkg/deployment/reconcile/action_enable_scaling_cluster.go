//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
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
	"github.com/rs/zerolog"
)

func init() {
	registerAction(api.ActionTypeEnableClusterScaling, newEnableScalingCluster)
}

// newEnableScalingCluster creates the new action with enabling scaling DBservers and coordinators.
func newEnableScalingCluster(log zerolog.Logger, action api.Action, actionCtx ActionContext) Action {
	a := &actionEnableScalingCluster{}

	a.actionImpl = newActionImpl(log, action, actionCtx, 0, util.NewString(""))

	return a
}

// actionEnableScalingCluster implements enabling scaling DBservers and coordinators.
type actionEnableScalingCluster struct {
	// actionImpl implement timeout and member id functions
	actionImpl

	// actionEmptyCheckProgress implement check progress with empty implementation
	actionEmptyCheckProgress
}

// Start enables scaling DBservers and coordinators
func (a *actionEnableScalingCluster) Start(ctx context.Context) (bool, error) {
	err := a.actionCtx.EnableScalingCluster()
	if err != nil {
		return false, err
	}
	return true, nil
}
