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
	"encoding/base64"

	"github.com/arangodb/kube-arangodb/pkg/util/globals"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"

	"github.com/arangodb/kube-arangodb/pkg/deployment/patch"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"

	"k8s.io/apimachinery/pkg/types"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/resources"
	"github.com/rs/zerolog"
)

func init() {
	registerAction(api.ActionTypeAppendTLSCACertificate, newAppendTLSCACertificateAction, operationTLSCACertificateTimeout)
}

func newAppendTLSCACertificateAction(log zerolog.Logger, action api.Action, actionCtx ActionContext) Action {
	a := &appendTLSCACertificateAction{}

	a.actionImpl = newActionImplDefRef(log, action, actionCtx)

	return a
}

type appendTLSCACertificateAction struct {
	actionImpl

	actionEmptyCheckProgress
}

func (a *appendTLSCACertificateAction) Start(ctx context.Context) (bool, error) {
	if !a.actionCtx.GetSpec().TLS.IsSecure() {
		return true, nil
	}

	certChecksum, exists := a.action.Params[checksum]
	if !exists {
		a.log.Warn().Msgf("Key %s is missing in action", checksum)
		return true, nil
	}

	caSecret, exists := a.actionCtx.ACS().CurrentClusterCache().Secret().V1().GetSimple(a.actionCtx.GetSpec().TLS.GetCASecretName())
	if !exists {
		a.log.Warn().Msgf("Secret %s is missing", a.actionCtx.GetSpec().TLS.GetCASecretName())
		return true, nil
	}

	caFolder, exists := a.actionCtx.ACS().CurrentClusterCache().Secret().V1().GetSimple(resources.GetCASecretName(a.actionCtx.GetAPIObject()))
	if !exists {
		a.log.Warn().Msgf("Secret %s is missing", resources.GetCASecretName(a.actionCtx.GetAPIObject()))
		return true, nil
	}

	ca, _, err := resources.GetKeyCertFromSecret(a.log, caSecret, resources.CACertName, resources.CAKeyName)
	if err != nil {
		a.log.Warn().Err(err).Msgf("Cert %s is invalid", resources.GetCASecretName(a.actionCtx.GetAPIObject()))
		return true, nil
	}

	caData, err := ca.ToPem()
	if err != nil {
		a.log.Warn().Err(err).Str("secret", resources.GetCASecretName(a.actionCtx.GetAPIObject())).Msgf("Unable to parse ca into pem")
		return true, nil
	}

	caSha := util.SHA256(caData)

	if caSha != certChecksum {
		a.log.Warn().Msgf("Cert changed")
		return true, nil
	}

	if _, exists := caFolder.Data[caSha]; exists {
		a.log.Warn().Msgf("Cert already exists")
		return true, nil
	}

	p := patch.NewPatch()
	p.ItemAdd(patch.NewPath("data", caSha), base64.StdEncoding.EncodeToString(caData))

	patch, err := p.Marshal()
	if err != nil {
		a.log.Error().Err(err).Msgf("Unable to encrypt patch")
		return true, nil
	}

	err = globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
		_, err := a.actionCtx.ACS().CurrentClusterCache().SecretsModInterface().V1().Patch(ctxChild, resources.GetCASecretName(a.actionCtx.GetAPIObject()), types.JSONPatchType, patch, meta.PatchOptions{})
		return err
	})
	if err != nil {
		if !k8sutil.IsInvalid(err) {
			return false, errors.Wrapf(err, "Unable to update secret: %s", string(patch))
		}
	}

	return true, nil
}
