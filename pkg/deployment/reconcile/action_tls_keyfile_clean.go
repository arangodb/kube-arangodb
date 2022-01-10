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

	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/rs/zerolog"
)

func init() {
	registerAction(api.ActionTypeCleanTLSKeyfileCertificate, newCleanTLSKeyfileCertificateAction)
}

func newCleanTLSKeyfileCertificateAction(log zerolog.Logger, action api.Action, actionCtx ActionContext) Action {
	a := &cleanTLSKeyfileCertificateAction{}

	a.actionImpl = newActionImplDefRef(log, action, actionCtx, operationTLSCACertificateTimeout)

	return a
}

type cleanTLSKeyfileCertificateAction struct {
	actionImpl

	actionEmptyCheckProgress
}

func (a *cleanTLSKeyfileCertificateAction) Start(ctx context.Context) (bool, error) {
	if !a.actionCtx.GetSpec().TLS.IsSecure() {
		return true, nil
	}

	member, exists := a.actionCtx.GetMemberStatusByID(a.action.MemberID)
	if !exists {
		a.log.Warn().Msgf("Member does not exist")
		return true, nil
	}

	if err := a.actionCtx.DeleteTLSKeyfile(ctx, a.action.Group, member); err != nil {
		a.log.Warn().Err(err).Msgf("Unable to remove keyfile")
		if !k8sutil.IsNotFound(err) {
			return false, err
		}
	}

	return true, nil
}
