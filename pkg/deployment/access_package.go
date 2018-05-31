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
	"github.com/ghodss/yaml"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

const (
	clientAuthValidFor = time.Hour * 24 * 365 // 1yr
	clientAuthCurve    = "P256"
)

// createAccessPackages creates a arangosync access packages specified
// in spec.sync.externalAccess.accessPackageSecretNames.
func (d *Deployment) createAccessPackages() error {
	spec := d.apiObject.Spec

	if !spec.Sync.IsEnabled() {
		// We're only relevant when sync is enabled
		return nil
	}

	for _, apSecretName := range spec.Sync.ExternalAccess.AccessPackageSecretNames {
		if err := d.ensureAccessPackage(apSecretName); err != nil {
			return maskAny(err)
		}
	}

	return nil
}

// ensureAccessPackage creates an arangosync access package with given name
// it is does not already exist.
func (d *Deployment) ensureAccessPackage(apSecretName string) error {
	log := d.deps.Log
	ns := d.GetNamespace()
	secrets := d.deps.KubeCli.CoreV1().Secrets(ns)
	spec := d.apiObject.Spec

	if _, err := secrets.Get(apSecretName, metav1.GetOptions{}); err == nil {
		// Secret already exists
		return nil
	}

	// Fetch client authentication CA
	clientAuthSecretName := spec.Sync.Authentication.GetClientCASecretName()
	clientAuthCert, clientAuthKey, err := k8sutil.GetCASecret(d.deps.KubeCli.CoreV1(), clientAuthSecretName, ns)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to get client-auth CA secret")
		return maskAny(err)
	}

	// Fetch TLS CA public key
	tlsCASecretName := spec.Sync.TLS.GetCASecretName()
	tlsCACert, err := k8sutil.GetCACertficateSecret(d.deps.KubeCli.CoreV1(), tlsCASecretName, ns)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to get TLS CA secret")
		return maskAny(err)
	}

	// Create keyfile
	ca, err := certificates.LoadCAFromPEM(clientAuthCert, clientAuthKey)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to parse client-auth CA")
		return maskAny(err)
	}

	// Create certificate
	options := certificates.CreateCertificateOptions{
		ValidFor:     clientAuthValidFor,
		ECDSACurve:   clientAuthCurve,
		IsClientAuth: true,
	}
	cert, key, err := certificates.CreateCertificate(options, &ca)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to create client-auth keyfile")
		return maskAny(err)
	}
	keyfile := strings.TrimSpace(cert) + "\n" + strings.TrimSpace(key)

	// Create secrets (in memory)
	keyfileSecret := v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: apSecretName + "-auth",
			Labels: map[string]string{
				"remote-deployment": d.apiObject.GetName(),
			},
		},
		Data: map[string][]byte{
			constants.SecretTLSKeyfile: []byte(keyfile),
		},
		Type: "Opaque",
	}
	tlsCASecret := v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: apSecretName + "-ca",
			Labels: map[string]string{
				"remote-deployment": d.apiObject.GetName(),
			},
		},
		Data: map[string][]byte{
			constants.SecretCACertificate: []byte(tlsCACert),
		},
		Type: "Opaque",
	}

	// Serialize secrets
	keyfileYaml, err := yaml.Marshal(keyfileSecret)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to encode client-auth keyfile Secret")
		return maskAny(err)
	}
	tlsCAYaml, err := yaml.Marshal(tlsCASecret)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to encode TLS CA Secret")
		return maskAny(err)
	}
	allYaml := strings.TrimSpace(string(keyfileYaml)) + "\n---\n" + strings.TrimSpace(string(tlsCAYaml))

	// Create secret containing access package
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: apSecretName,
		},
		Data: map[string][]byte{
			constants.SecretAccessPackageYaml: []byte(allYaml),
		},
	}
	// Attach secret to owner
	secret.SetOwnerReferences(append(secret.GetOwnerReferences(), d.apiObject.AsOwner()))
	if _, err := secrets.Create(secret); err != nil {
		// Failed to create secret
		log.Debug().Err(err).Str("secret-name", apSecretName).Msg("Failed to create access package Secret")
		return maskAny(err)
	}

	// Write log entry & create event
	log.Info().Str("secret-name", apSecretName).Msg("Created access package Secret")
	d.CreateEvent(k8sutil.NewAccessPackageCreatedEvent(d.apiObject, apSecretName))

	return nil
}
