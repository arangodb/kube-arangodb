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
	"k8s.io/apimachinery/pkg/types"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/patch"
	"github.com/arangodb/kube-arangodb/pkg/deployment/resources"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
)

func newCleanTLSCACertificateAction(action api.Action, actionCtx ActionContext) Action {
	a := &actionCleanTLSCACertificate{}

	a.actionImpl = newActionImplDefRef(action, actionCtx)

	return a
}

type actionCleanTLSCACertificate struct {
	actionImpl

	actionEmptyCheckProgress
}

func (a *actionCleanTLSCACertificate) Start(ctx context.Context) (bool, error) {
	a.log.Info("Clean TLS Ca")
	if !a.actionCtx.GetSpec().TLS.IsSecure() {
		a.log.Info("Insecure deployment")
		return true, nil
	}

	certChecksum, exists := a.action.Params[checksum]
	if !exists {
		a.log.Warn("Key %s is missing in action", checksum)
		return true, nil
	}

	caSecret, exists := a.actionCtx.ACS().CurrentClusterCache().Secret().V1().GetSimple(a.actionCtx.GetSpec().TLS.GetCASecretName())
	if !exists {
		a.log.Warn("Secret %s is missing", a.actionCtx.GetSpec().TLS.GetCASecretName())
		return true, nil
	}

	caFolder, exists := a.actionCtx.ACS().CurrentClusterCache().Secret().V1().GetSimple(resources.GetCASecretName(a.actionCtx.GetAPIObject()))
	if !exists {
		a.log.Warn("Secret %s is missing", resources.GetCASecretName(a.actionCtx.GetAPIObject()))
		return true, nil
	}

	ca, _, err := resources.GetKeyCertFromSecret(caSecret, resources.CACertName, resources.CAKeyName)
	if err != nil {
		a.log.Err(err).Warn("Cert %s is invalid", resources.GetCASecretName(a.actionCtx.GetAPIObject()))
		return true, nil
	}

	caData, err := ca.ToPem()
	if err != nil {
		a.log.Err(err).Str("secret", resources.GetCASecretName(a.actionCtx.GetAPIObject())).Warn("Unable to parse ca into pem")
		return true, nil
	}

	caSha := util.SHA256(caData)

	if caSha == certChecksum {
		a.log.Warn("Unable to remove current ca")
		return true, nil
	}

	if _, exists := caFolder.Data[certChecksum]; !exists {
		a.log.Warn("Cert missing")
		return true, nil
	}

	p := patch.NewPatch()
	p.ItemRemove(patch.NewPath("data", certChecksum))

	patch, err := p.Marshal()
	if err != nil {
		a.log.Err(err).Error("Unable to encrypt patch")
		return true, nil
	}

	a.log.Info("Removing key %s from truststore", certChecksum)

	err = globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
		_, err := a.actionCtx.ACS().CurrentClusterCache().SecretsModInterface().V1().Patch(ctxChild, resources.GetCASecretName(a.actionCtx.GetAPIObject()), types.JSONPatchType, patch, meta.PatchOptions{})
		return err
	})
	if err != nil {
		if !kerrors.IsInvalid(err) {
			return false, errors.Wrapf(err, "Unable to update secret: %s", string(patch))
		}
	}

	return true, nil
}
