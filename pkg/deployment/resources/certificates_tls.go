//
// DISCLAIMER
//
// Copyright 2018 ArangoDB GmbH, Cologne, Germany
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
// Author Ewout Prangsma
//

package resources

import (
	"fmt"
	"strings"
	"time"

	certificates "github.com/arangodb-helper/go-certificates"
	"github.com/rs/zerolog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/typed/core/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

const (
	caTTL         = time.Hour * 24 * 365 * 10 // 10 year
	tlsECDSACurve = "P256"                    // This curve is the default that ArangoDB accepts and plenty strong
)

// createTLSCACertificate creates a CA certificate and stores it in a secret with name
// specified in the given spec.
func createTLSCACertificate(log zerolog.Logger, secrets k8sutil.SecretInterface, spec api.TLSSpec, deploymentName string, ownerRef *metav1.OwnerReference) error {
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
		return maskAny(err)
	}
	if err := k8sutil.CreateCASecret(secrets, spec.GetCASecretName(), cert, priv, ownerRef); err != nil {
		if k8sutil.IsAlreadyExists(err) {
			log.Debug().Msg("CA Secret already exists")
		} else {
			log.Debug().Err(err).Msg("Failed to create CA Secret")
		}
		return maskAny(err)
	}
	log.Debug().Msg("Created CA Secret")
	return nil
}

// createTLSServerCertificate creates a TLS certificate for a specific server and stores
// it in a secret with the given name.
func createTLSServerCertificate(log zerolog.Logger, secrets v1.SecretInterface, serverNames []string, spec api.TLSSpec,
	secretName string, ownerRef *metav1.OwnerReference) error {

	log = log.With().Str("secret", secretName).Logger()
	// Load alt names
	dnsNames, ipAddresses, emailAddress, err := spec.GetParsedAltNames()
	if err != nil {
		log.Debug().Err(err).Msg("Failed to get alternate names")
		return maskAny(err)
	}

	// Load CA certificate
	caCert, caKey, _, err := k8sutil.GetCASecret(secrets, spec.GetCASecretName(), nil)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to load CA certificate")
		return maskAny(err)
	}
	ca, err := certificates.LoadCAFromPEM(caCert, caKey)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to decode CA certificate")
		return maskAny(err)
	}

	options := certificates.CreateCertificateOptions{
		CommonName:     serverNames[0],
		Hosts:          append(append(serverNames, dnsNames...), ipAddresses...),
		EmailAddresses: emailAddress,
		ValidFrom:      time.Now(),
		ValidFor:       spec.GetTTL().AsDuration(),
		IsCA:           false,
		ECDSACurve:     tlsECDSACurve,
	}
	cert, priv, err := certificates.CreateCertificate(options, &ca)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to create server certificate")
		return maskAny(err)
	}
	keyfile := strings.TrimSpace(cert) + "\n" +
		strings.TrimSpace(priv)
	if err := k8sutil.CreateTLSKeyfileSecret(secrets, secretName, keyfile, ownerRef); err != nil {
		if k8sutil.IsAlreadyExists(err) {
			log.Debug().Msg("Server Secret already exists")
		} else {
			log.Debug().Err(err).Msg("Failed to create server Secret")
		}
		return maskAny(err)
	}
	log.Debug().Msg("Created server Secret")
	return nil
}
