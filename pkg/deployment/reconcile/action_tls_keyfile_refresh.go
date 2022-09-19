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
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

func newRefreshTLSKeyfileCertificateAction(action api.Action, actionCtx ActionContext) Action {
	a := &actionRefreshTLSKeyfileCertificate{}

	a.actionImpl = newActionImplDefRef(action, actionCtx)

	return a
}

type actionRefreshTLSKeyfileCertificate struct {
	actionImpl
}

func (a *actionRefreshTLSKeyfileCertificate) CheckProgress(ctx context.Context) (bool, bool, error) {
	c, err := a.actionCtx.GetMembersState().GetMemberClient(a.action.MemberID)
	if err != nil {
		a.log.Err(err).Warn("Unable to get client")
		return true, false, nil
	}

	s, exists := a.actionCtx.ACS().CurrentClusterCache().Secret().V1().GetSimple(k8sutil.CreateTLSKeyfileSecretName(a.actionCtx.GetAPIObject().GetName(), a.action.Group.AsRole(), a.action.MemberID))
	if !exists {
		a.log.Warn("Keyfile secret is missing")
		return true, false, nil
	}

	keyfile, ok := s.Data[constants.SecretTLSKeyfile]
	if !ok {
		a.log.Warn("Keyfile secret is invalid")
		return true, false, nil
	}

	keyfileSha := util.SHA256(keyfile)

	client := client.NewClient(c.Connection(), a.log)

	ctxChild, cancel := globals.GetGlobalTimeouts().ArangoD().WithTimeout(ctx)
	defer cancel()
	e, err := client.RefreshTLS(ctxChild)
	if err != nil {
		a.log.Err(err).Warn("Unable to refresh TLS")
		return true, false, nil
	}

	if e.Result.KeyFile.GetSHA().Checksum() == keyfileSha {
		return true, false, nil
	}

	return false, false, nil
}

func (a *actionRefreshTLSKeyfileCertificate) Start(ctx context.Context) (bool, error) {
	ready, _, err := a.CheckProgress(ctx)
	return ready, err
}
