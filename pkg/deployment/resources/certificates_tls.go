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

package resources

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/arangodb/kube-arangodb/pkg/util/globals"

	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/tls"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"

	certificates "github.com/arangodb-helper/go-certificates"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	secretv1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/secret/v1"
	"github.com/rs/zerolog"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	caTTL         = time.Hour * 24 * 365 * 10 // 10 year
	tlsECDSACurve = "P256"                    // This curve is the default that ArangoDB accepts and plenty strong
)

// createTLSCACertificate creates a CA certificate and stores it in a secret with name
// specified in the given spec.
func createTLSCACertificate(ctx context.Context, log zerolog.Logger, secrets secretv1.ModInterface, spec api.TLSSpec,
	deploymentName string, ownerRef *meta.OwnerReference) error {
	log = log.With().Str("secret", spec.GetCASecretName()).Logger()

	options := certificates.CreateCertificateOptions{
		CommonName: fmt.Sprintf("%s Root Certificate", deploymentName),
		ValidFrom:  time.Now(),
		ValidFor:   caTTL,
		IsCA:       true,
		ECDSACurve: tlsECDSACurve,
	}
	cert, priv, err := certificates.CreateCertificate(options, nil)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to create CA certificate")
		return errors.WithStack(err)
	}
	if err := k8sutil.CreateCASecret(ctx, secrets, spec.GetCASecretName(), cert, priv, ownerRef); err != nil {
		if k8sutil.IsAlreadyExists(err) {
			log.Debug().Msg("CA Secret already exists")
		} else {
			log.Debug().Err(err).Msg("Failed to create CA Secret")
		}
		return errors.WithStack(err)
	}
	log.Debug().Msg("Created CA Secret")
	return nil
}

// createTLSServerCertificate creates a TLS certificate for a specific server and stores
// it in a secret with the given name.
func createTLSServerCertificate(ctx context.Context, log zerolog.Logger, cachedStatus inspectorInterface.Inspector, secrets secretv1.ModInterface, names tls.KeyfileInput, spec api.TLSSpec,
	secretName string, ownerRef *meta.OwnerReference) (bool, error) {
	log = log.With().Str("secret", secretName).Logger()
	// Load CA certificate
	ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
	defer cancel()
	caCert, caKey, _, err := k8sutil.GetCASecret(ctxChild, cachedStatus.Secret().V1().Read(), spec.GetCASecretName(), nil)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to load CA certificate")
		return false, errors.WithStack(err)
	}
	ca, err := certificates.LoadCAFromPEM(caCert, caKey)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to decode CA certificate")
		return false, errors.WithStack(err)
	}

	options := certificates.CreateCertificateOptions{
		CommonName:     names.AltNames[0],
		Hosts:          names.AltNames,
		EmailAddresses: names.Email,
		ValidFrom:      time.Now(),
		ValidFor:       spec.GetTTL().AsDuration(),
		IsCA:           false,
		ECDSACurve:     tlsECDSACurve,
	}
	cert, priv, err := certificates.CreateCertificate(options, &ca)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to create server certificate")
		return false, errors.WithStack(err)
	}
	keyfile := strings.TrimSpace(cert) + "\n" +
		strings.TrimSpace(priv)

	err = globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
		return k8sutil.CreateTLSKeyfileSecret(ctxChild, secrets, secretName, keyfile, ownerRef)
	})
	if err != nil {
		if k8sutil.IsAlreadyExists(err) {
			log.Debug().Msg("Server Secret already exists")
		} else {
			log.Debug().Err(err).Msg("Failed to create server Secret")
		}
		return false, errors.WithStack(err)
	}
	log.Debug().Msg("Created server Secret")
	return true, nil
}
