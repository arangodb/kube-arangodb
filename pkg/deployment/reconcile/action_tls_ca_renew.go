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
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/rs/zerolog"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {
	registerAction(api.ActionTypeRenewTLSCACertificate, newRenewTLSCACertificateAction)
}

func newRenewTLSCACertificateAction(log zerolog.Logger, action api.Action, actionCtx ActionContext) Action {
	a := &renewTLSCACertificateAction{}

	a.actionImpl = newActionImplDefRef(log, action, actionCtx, operationTLSCACertificateTimeout)

	return a
}

type renewTLSCACertificateAction struct {
	actionImpl

	actionEmptyCheckProgress
}

func (a *renewTLSCACertificateAction) Start(ctx context.Context) (bool, error) {
	if !a.actionCtx.GetSpec().TLS.IsSecure() {
		return true, nil
	}

	s := a.actionCtx.SecretsInterface()
	if err := s.Delete(a.actionCtx.GetSpec().TLS.GetCASecretName(), &meta.DeleteOptions{}); err != nil {
		if !k8sutil.IsNotFound(err) {
			a.log.Warn().Err(err).Msgf("Unable to clean cert %s", a.actionCtx.GetSpec().TLS.GetCASecretName())
			return true, nil
		}
	}

	return true, nil
}
