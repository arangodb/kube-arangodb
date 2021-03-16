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
	"github.com/arangodb/kube-arangodb/pkg/deployment/client"
	"github.com/arangodb/kube-arangodb/pkg/deployment/pod"
	"github.com/rs/zerolog"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {
	registerAction(api.ActionTypeEncryptionKeyRefresh, newEncryptionKeyRefresh)
}

func newEncryptionKeyRefresh(log zerolog.Logger, action api.Action, actionCtx ActionContext) Action {
	a := &encryptionKeyRefreshAction{}

	a.actionImpl = newActionImplDefRef(log, action, actionCtx, defaultTimeout)

	return a
}

type encryptionKeyRefreshAction struct {
	actionImpl
}

func (a *encryptionKeyRefreshAction) Start(ctx context.Context) (bool, error) {
	ready, _, err := a.CheckProgress(ctx)
	return ready, err
}

func (a *encryptionKeyRefreshAction) CheckProgress(ctx context.Context) (bool, bool, error) {
	keyfolder, err := a.actionCtx.SecretsInterface().Get(ctx, pod.GetEncryptionFolderSecretName(a.actionCtx.GetName()), meta.GetOptions{})
	if err != nil {
		a.log.Err(err).Msgf("Unable to fetch encryption folder")
		return true, false, nil
	}

	c, err := a.actionCtx.GetServerClient(ctx, a.action.Group, a.action.MemberID)
	if err != nil {
		a.log.Warn().Err(err).Msg("Unable to get client")
		return true, false, nil
	}

	client := client.NewClient(c.Connection())

	e, err := client.RefreshEncryption(ctx)
	if err != nil {
		a.log.Warn().Err(err).Msg("Unable to refresh encryption")
		return true, false, nil
	}

	if !e.Result.KeysPresent(keyfolder.Data) {
		return false, false, nil
	}

	return true, false, nil
}
