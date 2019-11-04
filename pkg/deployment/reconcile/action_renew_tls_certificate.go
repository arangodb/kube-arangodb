//
// DISCLAIMER
//
// Copyright 2018 ArangoDB GmbH, Cologne, Germany
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

// NewRenewTLSCertificateAction creates a new Action that implements the given
// planned RenewTLSCertificate action.
func NewRenewTLSCertificateAction(log zerolog.Logger, action api.Action, actionCtx ActionContext) Action {
	return &renewTLSCertificateAction{
		log:       log,
		action:    action,
		actionCtx: actionCtx,
	}
}

// renewTLSCertificateAction implements a RenewTLSCertificate action.
type renewTLSCertificateAction struct {
	log       zerolog.Logger
	action    api.Action
	actionCtx ActionContext
}

// Start performs the start of the action.
// Returns true if the action is completely finished, false in case
// the start time needs to be recorded and a ready condition needs to be checked.
func (a *renewTLSCertificateAction) Start(ctx context.Context) (bool, error) {
	log := a.log
	group := a.action.Group
	m, ok := a.actionCtx.GetMemberStatusByID(a.action.MemberID)
	if !ok {
		log.Error().Msg("No such member")
	}
	// Just delete the secret.
	// It will be re-created when the member restarts.
	if err := a.actionCtx.DeleteTLSKeyfile(group, m); err != nil {
		return false, maskAny(err)
	}
	return false, nil
}

// CheckProgress checks the progress of the action.
// Returns true if the action is completely finished, false otherwise.
func (a *renewTLSCertificateAction) CheckProgress(ctx context.Context) (bool, bool, error) {
	return true, false, nil
}

// Timeout returns the amount of time after which this action will timeout.
func (a *renewTLSCertificateAction) Timeout() time.Duration {
	return renewTLSCertificateTimeout
}

// Return the MemberID used / created in this action
func (a *renewTLSCertificateAction) MemberID() string {
	return a.action.MemberID
}
