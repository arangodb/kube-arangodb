//
// DISCLAIMER
//
// Copyright 2020-2021 ArangoDB GmbH, Cologne, Germany
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
// Author Tomasz Mielech
//

package reconcile

import (
	"context"

	"github.com/arangodb/kube-arangodb/pkg/util/globals"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/client"
	"github.com/arangodb/kube-arangodb/pkg/deployment/pod"
	"github.com/rs/zerolog"
)

func init() {
	registerAction(api.ActionTypeJWTRefresh, newJWTRefresh)
}

func newJWTRefresh(log zerolog.Logger, action api.Action, actionCtx ActionContext) Action {
	a := &jwtRefreshAction{}

	a.actionImpl = newActionImplDefRef(log, action, actionCtx, defaultTimeout)

	return a
}

type jwtRefreshAction struct {
	actionImpl
}

func (a *jwtRefreshAction) CheckProgress(ctx context.Context) (bool, bool, error) {
	if folder, err := ensureJWTFolderSupport(a.actionCtx.GetSpec(), a.actionCtx.GetStatusSnapshot()); err != nil || !folder {
		return true, false, nil
	}

	folder, ok := a.actionCtx.GetCachedStatus().Secret(pod.JWTSecretFolder(a.actionCtx.GetAPIObject().GetName()))
	if !ok {
		a.log.Error().Msgf("Unable to get JWT folder info")
		return true, false, nil
	}

	ctxChild, cancel := globals.GetGlobalTimeouts().ArangoD().WithTimeout(ctx)
	defer cancel()
	c, err := a.actionCtx.GetServerClient(ctxChild, a.action.Group, a.action.MemberID)
	if err != nil {
		a.log.Warn().Err(err).Msg("Unable to get client")
		return true, false, nil
	}

	ctxChild, cancel = globals.GetGlobalTimeouts().ArangoD().WithTimeout(ctx)
	defer cancel()
	if invalid, err := isMemberJWTTokenInvalid(ctxChild, client.NewClient(c.Connection()), folder.Data, true); err != nil {
		a.log.Warn().Err(err).Msg("Error while getting JWT Status")
		return true, false, nil
	} else if invalid {
		return false, false, nil
	}
	return true, false, nil
}

func (a *jwtRefreshAction) Start(ctx context.Context) (bool, error) {
	ready, _, err := a.CheckProgress(ctx)
	return ready, err
}
