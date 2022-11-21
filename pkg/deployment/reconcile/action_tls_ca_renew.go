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
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
)

func newRenewTLSCACertificateAction(action api.Action, actionCtx ActionContext) Action {
	a := &actionRenewTLSCACertificate{}

	a.actionImpl = newActionImplDefRef(action, actionCtx)

	return a
}

type actionRenewTLSCACertificate struct {
	actionImpl

	actionEmptyCheckProgress
}

func (a *actionRenewTLSCACertificate) Start(ctx context.Context) (bool, error) {
	if !a.actionCtx.GetSpec().TLS.IsSecure() {
		return true, nil
	}

	err := globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
		return a.actionCtx.ACS().CurrentClusterCache().SecretsModInterface().V1().Delete(ctxChild, a.actionCtx.GetSpec().TLS.GetCASecretName(), meta.DeleteOptions{})
	})
	if err != nil {
		if !kerrors.IsNotFound(err) {
			a.log.Err(err).Warn("Unable to clean cert %s", a.actionCtx.GetSpec().TLS.GetCASecretName())
			return true, nil
		}
	}

	return true, nil
}
