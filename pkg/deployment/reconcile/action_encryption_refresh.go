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

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/client"
	"github.com/arangodb/kube-arangodb/pkg/deployment/pod"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
)

func newEncryptionKeyRefreshAction(action api.Action, actionCtx ActionContext) Action {
	a := &actionEncryptionKeyRefresh{}

	a.actionImpl = newActionImplDefRef(action, actionCtx)

	return a
}

type actionEncryptionKeyRefresh struct {
	actionImpl
}

func (a *actionEncryptionKeyRefresh) Start(ctx context.Context) (bool, error) {
	ready, _, err := a.CheckProgress(ctx)
	return ready, err
}

func (a *actionEncryptionKeyRefresh) CheckProgress(ctx context.Context) (bool, bool, error) {
	ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
	defer cancel()
	keyFolder, err := a.actionCtx.ACS().CurrentClusterCache().Secret().V1().Read().Get(ctxChild,
		pod.GetEncryptionFolderSecretName(a.actionCtx.GetName()), meta.GetOptions{})
	if err != nil {
		a.log.Err(err).Error("Unable to fetch encryption folder")
		return true, false, nil
	}

	c, err := a.actionCtx.GetMembersState().GetMemberClient(a.action.MemberID)
	if err != nil {
		a.log.Err(err).Warn("Unable to get client")
		return true, false, nil
	}

	client := client.NewClient(c.Connection(), a.log)
	ctxChild, cancel = globals.GetGlobalTimeouts().ArangoD().WithTimeout(ctx)
	defer cancel()
	e, err := client.RefreshEncryption(ctxChild)
	if err != nil {
		a.log.Err(err).Warn("Unable to refresh encryption")
		return true, false, nil
	}

	if !e.Result.KeysPresent(keyFolder.Data) {
		return false, false, nil
	}

	return true, false, nil
}
