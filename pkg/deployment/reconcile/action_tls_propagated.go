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
// Author Adam Janikowski
//

package reconcile

import (
	"context"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/rs/zerolog"
)

func init() {
	registerAction(api.ActionTypeTLSPropagated, newTLSPropagated)
}

func newTLSPropagated(log zerolog.Logger, action api.Action, actionCtx ActionContext) Action {
	a := &tlsPropagatedAction{}

	a.actionImpl = newActionImplDefRef(log, action, actionCtx, defaultTimeout)

	return a
}

type tlsPropagatedAction struct {
	actionImpl

	actionEmptyCheckProgress
}

func (a *tlsPropagatedAction) Start(ctx context.Context) (bool, error) {
	propagatedFlag, exists := a.action.Params[propagated]
	if !exists {
		a.log.Error().Msgf("Propagated flag is missing")
		return true, nil
	}

	propagatedFlagBool := propagatedFlag == conditionTrue

	if err := a.actionCtx.WithStatusUpdate(func(s *api.DeploymentStatus) bool {
		if s.Hashes.TLS.Propagated != propagatedFlagBool {
			s.Hashes.TLS.Propagated = propagatedFlagBool
			return true
		}

		return false
	}); err != nil {
		return false, err
	}

	return true, nil
}
