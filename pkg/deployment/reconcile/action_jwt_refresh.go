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

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/client"
	"github.com/arangodb/kube-arangodb/pkg/deployment/pod"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
)

func newJWTRefreshAction(action api.Action, actionCtx ActionContext) Action {
	a := &actionJWTRefresh{}

	a.actionImpl = newActionImplDefRef(action, actionCtx)

	return a
}

type actionJWTRefresh struct {
	actionImpl
}

func (a *actionJWTRefresh) CheckProgress(ctx context.Context) (bool, bool, error) {
	if folder, err := ensureJWTFolderSupport(a.actionCtx.GetSpec(), a.actionCtx.GetStatus()); err != nil || !folder {
		return true, false, nil
	}

	folder, ok := a.actionCtx.ACS().CurrentClusterCache().Secret().V1().GetSimple(pod.JWTSecretFolder(a.actionCtx.GetAPIObject().GetName()))
	if !ok {
		a.log.Error("Unable to get JWT folder info")
		return true, false, nil
	}

	c, err := a.actionCtx.GetMembersState().GetMemberClient(a.action.MemberID)
	if err != nil {
		a.log.Err(err).Warn("Unable to get client")
		return true, false, nil
	}

	ctxChild, cancel := globals.GetGlobalTimeouts().ArangoD().WithTimeout(ctx)
	defer cancel()
	if invalid, err := isMemberJWTTokenInvalid(ctxChild, client.NewClient(c.Connection(), a.log), folder.Data, true); err != nil {
		a.log.Err(err).Warn("Error while getting JWT Status")
		return true, false, nil
	} else if invalid {
		return false, false, nil
	}
	return true, false, nil
}

func (a *actionJWTRefresh) Start(ctx context.Context) (bool, error) {
	ready, _, err := a.CheckProgress(ctx)
	return ready, err
}
