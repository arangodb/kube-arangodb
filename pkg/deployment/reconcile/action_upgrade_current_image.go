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
// Author Ewout Prangsma
//

package reconcile

import (
	"context"
	"time"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/rs/zerolog"
)

// NewSetCurrentImageAction creates a new Action that implements the given
// planned SetCurrentImage action.
func NewSetCurrentImageAction(log zerolog.Logger, action api.Action, actionCtx ActionContext) Action {
	return &setCurrentImageAction{
		log:       log,
		action:    action,
		actionCtx: actionCtx,
	}
}

// setCurrentImageAction implements an SetCurrentImage.
type setCurrentImageAction struct {
	log       zerolog.Logger
	action    api.Action
	actionCtx ActionContext
}

// Start performs the start of the action.
// Returns true if the action is completely finished, false in case
// the start time needs to be recorded and a ready condition needs to be checked.
func (a *setCurrentImageAction) Start(ctx context.Context) (bool, error) {
	ready, _, err := a.CheckProgress(ctx)
	if err != nil {
		return false, maskAny(err)
	}
	return ready, nil
}

// CheckProgress checks the progress of the action.
// Returns true if the action is completely finished, false otherwise.
func (a *setCurrentImageAction) CheckProgress(ctx context.Context) (bool, bool, error) {
	log := a.log

	imageInfo, found := a.actionCtx.GetImageInfo(a.action.Image)
	if !found {
		return false, false, nil
	}
	if err := a.actionCtx.SetCurrentImage(imageInfo); err != nil {
		return false, false, maskAny(err)
	}
	log.Info().Str("image", a.action.Image).Msg("Changed current image")
	return true, false, nil
}

// Timeout returns the amount of time after which this action will timeout.
func (a *setCurrentImageAction) Timeout() time.Duration {
	return upgradeMemberTimeout
}

// Return the MemberID used / created in this action
func (a *setCurrentImageAction) MemberID() string {
	return ""
}
