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
	"time"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/rs/zerolog"
)

// actionDisableScalingCluster implements disabling scaling DBservers and coordinators.
type actionDisableScalingCluster struct {
	log         zerolog.Logger
	action      api.Action
	actionCtx   ActionContext
	newMemberID string
}

// NewDisableScalingCluster creates the new action with disabling scaling DBservers and coordinators.
func NewDisableScalingCluster(log zerolog.Logger, action api.Action, actionCtx ActionContext) Action {
	return &actionDisableScalingCluster{
		log:       log,
		action:    action,
		actionCtx: actionCtx,
	}
}

// Start disables scaling DBservers and coordinators
func (a *actionDisableScalingCluster) Start(ctx context.Context) (bool, error) {
	err := a.actionCtx.DisableScalingCluster()
	if err != nil {
		return false, err
	}
	return true, nil
}

// CheckProgress does not matter. Everything is done in 'Start' function
func (a *actionDisableScalingCluster) CheckProgress(ctx context.Context) (bool, bool, error) {
	return true, false, nil
}

// Timeout does not matter. Everything is done in 'Start' function
func (a *actionDisableScalingCluster) Timeout() time.Duration {
	return 0
}

// MemberID is not used
func (a *actionDisableScalingCluster) MemberID() string {
	return ""
}
