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

	"github.com/arangodb/kube-arangodb/pkg/deployment/client"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"

	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/rs/zerolog"
)

func init() {
	registerAction(api.ActionTypeRefreshTLSKeyfileCertificate, newRefreshTLSKeyfileCertificateAction)
}

func newRefreshTLSKeyfileCertificateAction(log zerolog.Logger, action api.Action, actionCtx ActionContext) Action {
	a := &refreshTLSKeyfileCertificateAction{}

	a.actionImpl = newActionImplDefRef(log, action, actionCtx, operationTLSCACertificateTimeout)

	return a
}

type refreshTLSKeyfileCertificateAction struct {
	actionImpl
}

func (a *refreshTLSKeyfileCertificateAction) CheckProgress(ctx context.Context) (bool, bool, error) {
	ctxChild, cancel := globals.GetGlobalTimeouts().ArangoD().WithTimeout(ctx)
	defer cancel()
	c, err := a.actionCtx.GetServerClient(ctxChild, a.action.Group, a.action.MemberID)
	if err != nil {
		a.log.Warn().Err(err).Msg("Unable to get client")
		return true, false, nil
	}

	s, exists := a.actionCtx.GetCachedStatus().Secret(k8sutil.CreateTLSKeyfileSecretName(a.actionCtx.GetAPIObject().GetName(), a.action.Group.AsRole(), a.action.MemberID))
	if !exists {
		a.log.Warn().Msg("Keyfile secret is missing")
		return true, false, nil
	}

	keyfile, ok := s.Data[constants.SecretTLSKeyfile]
	if !ok {
		a.log.Warn().Msg("Keyfile secret is invalid")
		return true, false, nil
	}

	keyfileSha := util.SHA256(keyfile)

	client := client.NewClient(c.Connection())

	ctxChild, cancel = globals.GetGlobalTimeouts().ArangoD().WithTimeout(ctx)
	defer cancel()
	e, err := client.RefreshTLS(ctxChild)
	if err != nil {
		a.log.Warn().Err(err).Msg("Unable to refresh TLS")
		return true, false, nil
	}

	if e.Result.KeyFile.GetSHA().Checksum() == keyfileSha {
		return true, false, nil
	}

	return false, false, nil
}

func (a *refreshTLSKeyfileCertificateAction) Start(ctx context.Context) (bool, error) {
	ready, _, err := a.CheckProgress(ctx)
	return ready, err
}
