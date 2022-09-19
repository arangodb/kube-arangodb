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
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
)

func newUpdateTLSSNIAction(action api.Action, actionCtx ActionContext) Action {
	a := &actionUpdateTLSSNI{}

	a.actionImpl = newActionImplDefRef(action, actionCtx)

	return a
}

type actionUpdateTLSSNI struct {
	actionImpl

	actionEmptyStart
}

func (t *actionUpdateTLSSNI) CheckProgress(ctx context.Context) (bool, bool, error) {
	spec := t.actionCtx.GetSpec()
	if !spec.TLS.IsSecure() {
		return true, false, nil
	}

	if i, ok := t.actionCtx.GetCurrentImageInfo(); !ok || !i.Enterprise {
		return true, false, nil
	}

	sni := spec.TLS.SNI
	if sni == nil {
		return true, false, nil
	}

	fetchedSecrets, err := mapTLSSNIConfig(*sni, t.actionCtx.ACS().CurrentClusterCache())
	if err != nil {
		t.log.Err(err).Warn("Unable to get SNI desired state")
		return true, false, nil
	}

	c, err := t.actionCtx.GetMembersState().GetMemberClient(t.action.MemberID)
	if err != nil {
		t.log.Err(err).Warn("Unable to get client")
		return true, false, nil
	}

	ctxChild, cancel := globals.GetGlobalTimeouts().ArangoD().WithTimeout(ctx)
	defer cancel()
	if ok, err := compareTLSSNIConfig(ctxChild, t.log, c.Connection(), fetchedSecrets, true); err != nil {
		t.log.Err(err).Warn("Unable to compare TLS config")
		return true, false, nil
	} else {
		return ok, false, nil
	}
}
