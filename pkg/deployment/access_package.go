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
// Author Ewout Prangsma
//

package deployment

import (
	"context"
	"strings"
	"time"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"

	certificates "github.com/arangodb-helper/go-certificates"
	"github.com/ghodss/yaml"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

const (
	clientAuthValidFor         = time.Hour * 24 * 365 // 1yr
	clientAuthCurve            = "P256"
	labelKeyOriginalDeployment = "original-deployment-name"
)

// createAccessPackages creates a arangosync access packages specified
// in spec.sync.externalAccess.accessPackageSecretNames.
func (d *Deployment) createAccessPackages() error {
	log := d.deps.Log
	spec := d.apiObject.Spec
	secrets := d.deps.KubeCli.CoreV1().Secrets(d.GetNamespace())

	if !spec.Sync.IsEnabled() {
		// We're only relevant when sync is enabled
		return nil
	}

	// Create all access packages that we're asked to build
	apNameMap := make(map[string]struct{})
	for _, apSecretName := range spec.Sync.ExternalAccess.AccessPackageSecretNames {
		apNameMap[apSecretName] = struct{}{}
		if err := d.ensureAccessPackage(apSecretName); err != nil {
			return errors.WithStack(err)
		}
	}

	// Remove all access packages that we did build, but are no longer needed
	secretList, err := secrets.List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Debug().Err(err).Msg("Failed to list secrets")
		return errors.WithStack(err)
	}
	for _, secret := range secretList.Items {
		if d.isOwnerOf(&secret) {
			if _, found := secret.Data[constants.SecretAccessPackageYaml]; found {
				// Secret is an access package
				if _, wanted := apNameMap[secret.GetName()]; !wanted {
					// We found an obsolete access package secret. Remove it.
					if err := secrets.Delete(context.Background(), secret.GetName(), metav1.DeleteOptions{
						Preconditions: &metav1.Preconditions{UID: &secret.UID},
					}); err != nil && !k8sutil.IsNotFound(err) {
						// Not serious enough to stop everything now, just log and create an event
						log.Warn().Err(err).Msg("Failed to remove obsolete access package secret")
						d.CreateEvent(k8sutil.NewErrorEvent("Access Package cleanup failed", err, d.apiObject))
					} else {
						// Access package removed, notify user
						log.Info().Str("secret-name", secret.GetName()).Msg("Removed access package Secret")
						d.CreateEvent(k8sutil.NewAccessPackageDeletedEvent(d.apiObject, secret.GetName()))
					}
				}
			}
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

	if _, err := secrets.Get(context.Background(), apSecretName, metav1.GetOptions{}); err == nil {
		// Secret already exists
		return nil
	}

	// Fetch client authentication CA
	clientAuthSecretName := spec.Sync.Authentication.GetClientCASecretName()
	clientAuthCert, clientAuthKey, _, err := k8sutil.GetCASecret(secrets, clientAuthSecretName, nil)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to get client-auth CA secret")
		return errors.WithStack(err)
	}

	// Fetch TLS CA public key
	tlsCASecretName := spec.Sync.TLS.GetCASecretName()
	tlsCACert, err := k8sutil.GetCACertficateSecret(secrets, tlsCASecretName)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to get TLS CA secret")
		return errors.WithStack(err)
	}

	// Create keyfile
	ca, err := certificates.LoadCAFromPEM(clientAuthCert, clientAuthKey)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to parse client-auth CA")
		return errors.WithStack(err)
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
		return errors.WithStack(err)
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
				labelKeyOriginalDeployment: d.apiObject.GetName(),
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
				labelKeyOriginalDeployment: d.apiObject.GetName(),
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
		return errors.WithStack(err)
	}
	tlsCAYaml, err := yaml.Marshal(tlsCASecret)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to encode TLS CA Secret")
		return errors.WithStack(err)
	}
	allYaml := strings.TrimSpace(string(keyfileYaml)) + "\n---\n" + strings.TrimSpace(string(tlsCAYaml))

	// Create secret containing access package
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: apSecretName,
		},
		Data: map[string][]byte{
			constants.SecretAccessPackageYaml: []byte(allYaml),
			constants.SecretCACertificate:     []byte(tlsCACert),
			constants.SecretTLSKeyfile:        []byte(keyfile),
		},
	}
	// Attach secret to owner
	secret.SetOwnerReferences(append(secret.GetOwnerReferences(), d.apiObject.AsOwner()))
	if _, err := secrets.Create(context.Background(), secret, metav1.CreateOptions{}); err != nil {
		// Failed to create secret
		log.Debug().Err(err).Str("secret-name", apSecretName).Msg("Failed to create access package Secret")
		return errors.WithStack(err)
	}

	// Write log entry & create event
	log.Info().Str("secret-name", apSecretName).Msg("Created access package Secret")
	d.CreateEvent(k8sutil.NewAccessPackageCreatedEvent(d.apiObject, apSecretName))

	return nil
}
