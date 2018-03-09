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

package deployment

import (
	"strings"
	"time"

	certificates "github.com/arangodb-helper/go-certificates"
	"github.com/rs/zerolog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/typed/core/v1"

	api "github.com/arangodb/k8s-operator/pkg/apis/arangodb/v1alpha"
	"github.com/arangodb/k8s-operator/pkg/util/k8sutil"
)

const (
	caTTL             = time.Hour * 24 * 365 * 10 // 10 year
	defaultECDSACurve = "P256"
)

// createCACertificate creates a CA certificate and stores it in a secret with name
// specified in the given spec.
func createCACertificate(log zerolog.Logger, cli v1.CoreV1Interface, spec api.TLSSpec, namespace string, ownerRef *metav1.OwnerReference) error {
	dnsNames, ipAddresses, emailAddress, err := spec.GetAltNames()
	if err != nil {
		log.Debug().Err(err).Msg("Failed to get alternate names")
		return maskAny(err)
	}

	options := certificates.CreateCertificateOptions{
		Hosts:          append(dnsNames, ipAddresses...),
		EmailAddresses: emailAddress,
		ValidFrom:      time.Now(),
		ValidFor:       caTTL,
		IsCA:           true,
		ECDSACurve:     defaultECDSACurve,
	}
	cert, priv, err := certificates.CreateCertificate(options, nil)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to create CA certificate")
		return maskAny(err)
	}
	if err := k8sutil.CreateCASecret(cli, spec.CASecretName, namespace, cert, priv, ownerRef); err != nil {
		log.Debug().Err(err).Msg("Failed to create CA Secret")
		return maskAny(err)
	}
	log.Debug().Str("secret", spec.CASecretName).Msg("Created CA Secret")
	return nil
}

// createServerCertificate creates a TLS certificate for a specific server and stores
// it in a secret with the given name.
func createServerCertificate(log zerolog.Logger, cli v1.CoreV1Interface, serverNames []string, spec api.TLSSpec, secretName, namespace string, ownerRef *metav1.OwnerReference) error {
	// Load alt names
	dnsNames, ipAddresses, emailAddress, err := spec.GetAltNames()
	if err != nil {
		log.Debug().Err(err).Msg("Failed to get alternate names")
		return maskAny(err)
	}

	// Load CA certificate
	caCert, caKey, err := k8sutil.GetCASecret(cli, spec.CASecretName, namespace)
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
		Hosts:          append(append(serverNames, dnsNames...), ipAddresses...),
		EmailAddresses: emailAddress,
		ValidFrom:      time.Now(),
		ValidFor:       spec.TTL,
		IsCA:           false,
		ECDSACurve:     defaultECDSACurve,
	}
	cert, priv, err := certificates.CreateCertificate(options, &ca)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to create server certificate")
		return maskAny(err)
	}
	keyfile := strings.TrimSpace(cert) + "\n" +
		strings.TrimSpace(priv)
	if err := k8sutil.CreateTLSKeyfileSecret(cli, secretName, namespace, keyfile, ownerRef); err != nil {
		log.Debug().Err(err).Msg("Failed to create server Secret")
		return maskAny(err)
	}
	log.Debug().Str("secret", secretName).Msg("Created server Secret")
	return nil
}
