//
// DISCLAIMER
//
// Copyright 2016-2024 ArangoDB GmbH, Cologne, Germany
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

package resources

import (
	"context"
	"fmt"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/logging"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
	secretv1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/secret/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
	ktls "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/tls"
)

// createTLSCACertificate creates a CA certificate and stores it in a secret with name
// specified in the given spec.
func (r *Resources) createTLSCACertificate(ctx context.Context, secrets secretv1.ModInterface, spec api.TLSSpec,
	deploymentName string, ownerRef *meta.OwnerReference) error {
	log := r.log.Str("section", "tls").Str("secret", spec.GetCASecretName())

	cert, priv, err := ktls.CreateTLSCACertificate(fmt.Sprintf("%s Root Certificate", deploymentName))
	if err != nil {
		log.Err(err).Debug("Failed to create CA certificate")
		return errors.WithStack(err)
	}

	if err := k8sutil.CreateCASecret(ctx, secrets, spec.GetCASecretName(), cert, priv, ownerRef); err != nil {
		if kerrors.IsAlreadyExists(err) {
			log.Debug("CA Secret already exists")
		} else {
			log.Err(err).Debug("Failed to create CA Secret")
		}
		return errors.WithStack(err)
	}
	log.Debug("Created CA Secret")
	return nil
}

// createTLSServerCertificate creates a TLS certificate for a specific server and stores
// it in a secret with the given name.
func createTLSServerCertificate(ctx context.Context, log logging.Logger, cachedStatus inspectorInterface.Inspector, secrets secretv1.ModInterface, names ktls.KeyfileInput, spec api.TLSSpec,
	secretName string, ownerRef *meta.OwnerReference) (bool, error) {
	log = log.Str("secret", secretName)
	// Setup defaults
	if names.TTL == nil {
		names.TTL = util.NewType(spec.GetTTL().AsDuration())
	}
	// Load CA certificate
	ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
	defer cancel()
	caCert, caKey, _, err := k8sutil.GetCASecret(ctxChild, cachedStatus.Secret().V1().Read(), spec.GetCASecretName(), nil)
	if err != nil {
		log.Err(err).Debug("Failed to load CA certificate")
		return false, errors.WithStack(err)
	}
	keyfile, err := ktls.CreateTLSServerKeyfile(caCert, caKey, names)
	if err != nil {
		log.Err(err).Debug("Failed to create server certificate")
		return false, errors.WithStack(err)
	}

	err = globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
		_, err := k8sutil.CreateTLSKeyfileSecret(ctxChild, secrets, secretName, keyfile, ownerRef)
		return err
	})
	if err != nil {
		if kerrors.IsAlreadyExists(err) {
			log.Debug("Server Secret already exists")
		} else {
			log.Err(err).Debug("Failed to create server Secret")
		}
		return false, errors.WithStack(err)
	}
	log.Debug("Created server Secret")
	return true, nil
}
