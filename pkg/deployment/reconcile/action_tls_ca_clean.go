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

	"github.com/arangodb/kube-arangodb/pkg/deployment/patch"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/resources"
	"github.com/rs/zerolog"
)

func init() {
	registerAction(api.ActionTypeCleanTLSCACertificate, newCleanTLSCACertificateAction)
}

func newCleanTLSCACertificateAction(log zerolog.Logger, action api.Action, actionCtx ActionContext) Action {
	a := &cleanTLSCACertificateAction{}

	a.actionImpl = newActionImplDefRef(log, action, actionCtx, operationTLSCACertificateTimeout)

	return a
}

type cleanTLSCACertificateAction struct {
	actionImpl

	actionEmptyCheckProgress
}

func (a *cleanTLSCACertificateAction) Start(ctx context.Context) (bool, error) {
	a.log.Info().Msgf("Clean TLS Ca")
	if !a.actionCtx.GetSpec().TLS.IsSecure() {
		a.log.Info().Msgf("Insecure deployment")
		return true, nil
	}

	certChecksum, exists := a.action.Params[checksum]
	if !exists {
		a.log.Warn().Msgf("Key %s is missing in action", checksum)
		return true, nil
	}

	caSecret, exists := a.actionCtx.GetCachedStatus().Secret(a.actionCtx.GetSpec().TLS.GetCASecretName())
	if !exists {
		a.log.Warn().Msgf("Secret %s is missing", a.actionCtx.GetSpec().TLS.GetCASecretName())
		return true, nil
	}

	caFolder, exists := a.actionCtx.GetCachedStatus().Secret(resources.GetCASecretName(a.actionCtx.GetAPIObject()))
	if !exists {
		a.log.Warn().Msgf("Secret %s is missing", resources.GetCASecretName(a.actionCtx.GetAPIObject()))
		return true, nil
	}

	ca, _, err := getKeyCertFromSecret(a.log, caSecret, resources.CACertName, resources.CAKeyName)
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

	if caSha == certChecksum {
		a.log.Warn().Msgf("Unable to remove current ca")
		return true, nil
	}

	if _, exists := caFolder.Data[certChecksum]; !exists {
		a.log.Warn().Msgf("Cert missing")
		return true, nil
	}

	p := patch.NewPatch()
	p.ItemRemove(patch.NewPath("data", certChecksum))

	patch, err := p.Marshal()
	if err != nil {
		a.log.Error().Err(err).Msgf("Unable to encrypt patch")
		return true, nil
	}

	a.log.Info().Msgf("Removing key %s from truststore", certChecksum)
	_, err = a.actionCtx.SecretsInterface().Patch(resources.GetCASecretName(a.actionCtx.GetAPIObject()), types.JSONPatchType, patch)
	if err != nil {
		if !k8sutil.IsInvalid(err) {
			return false, errors.Wrapf(err, "Unable to update secret: %s", string(patch))
		}
	}

	return true, nil
}
