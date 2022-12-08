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

package deployment

import (
	"context"
	"strings"
	"time"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	certificates "github.com/arangodb-helper/go-certificates"

	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
)

const (
	clientAuthValidFor         = time.Hour * 24 * 365 // 1yr
	clientAuthCurve            = "P256"
	labelKeyOriginalDeployment = "original-deployment-name"
)

// createAccessPackages creates a arangosync access packages specified
// in spec.sync.externalAccess.accessPackageSecretNames.
func (d *Deployment) createAccessPackages(ctx context.Context) error {
	log := d.sectionLogger("access-package")
	spec := d.GetSpec()

	if !d.IsSyncEnabled() {
		// We're only relevant when sync is enabled
		return nil
	}

	// Create all access packages that we're asked to build
	apNameMap := make(map[string]struct{})
	for _, apSecretName := range spec.Sync.ExternalAccess.AccessPackageSecretNames {
		apNameMap[apSecretName] = struct{}{}
		if err := d.ensureAccessPackage(ctx, apSecretName); err != nil {
			return errors.WithStack(err)
		}
	}

	// Remove all access packages that we did build, but are no longer needed
	secretList := d.acs.CurrentClusterCache().Secret().V1().ListSimple()
	for _, secret := range secretList {
		if d.isOwnerOf(secret) {
			if _, found := secret.Data[constants.SecretAccessPackageYaml]; found {
				// Secret is an access package
				if _, wanted := apNameMap[secret.GetName()]; !wanted {
					// We found an obsolete access package secret. Remove it.
					err := globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
						return d.SecretsModInterface().Delete(ctxChild, secret.GetName(), meta.DeleteOptions{
							Preconditions: &meta.Preconditions{UID: &secret.UID},
						})
					})
					if err != nil && !kerrors.IsNotFound(err) {
						// Not serious enough to stop everything now, just sectionLogger and create an event
						log.Err(err).Warn("Failed to remove obsolete access package secret")
						d.CreateEvent(k8sutil.NewErrorEvent("Access Package cleanup failed", err, d.currentObject))
					} else {
						// Access package removed, notify user
						log.Str("secret-name", secret.GetName()).Info("Removed access package Secret")
						d.CreateEvent(k8sutil.NewAccessPackageDeletedEvent(d.currentObject, secret.GetName()))
					}
				}
			}
		}
	}

	return nil
}

// ensureAccessPackage creates an arangosync access package with given name
// it is does not already exist.
func (d *Deployment) ensureAccessPackage(ctx context.Context, apSecretName string) error {
	log := d.sectionLogger("access-package")
	spec := d.GetSpec()

	_, err := d.acs.CurrentClusterCache().Secret().V1().Read().Get(ctx, apSecretName, meta.GetOptions{})
	if err == nil {
		// Secret already exists
		return nil
	} else if !kerrors.IsNotFound(err) {
		log.Err(err).Str("name", apSecretName).Debug("Failed to get arangosync access package secret")
		return errors.WithStack(err)
	}

	// Fetch client authentication CA
	clientAuthSecretName := spec.Sync.Authentication.GetClientCASecretName()
	clientAuthCert, clientAuthKey, _, err := k8sutil.GetCASecret(ctx, d.acs.CurrentClusterCache().Secret().V1().Read(), clientAuthSecretName, nil)
	if err != nil {
		log.Err(err).Debug("Failed to get client-auth CA secret")
		return errors.WithStack(err)
	}

	// Fetch TLS CA public key
	tlsCASecretName := spec.Sync.TLS.GetCASecretName()
	tlsCACert, err := k8sutil.GetCACertficateSecret(ctx, d.acs.CurrentClusterCache().Secret().V1().Read(), tlsCASecretName)
	if err != nil {
		log.Err(err).Debug("Failed to get TLS CA secret")
		return errors.WithStack(err)
	}

	// Create keyfile
	ca, err := certificates.LoadCAFromPEM(clientAuthCert, clientAuthKey)
	if err != nil {
		log.Err(err).Debug("Failed to parse client-auth CA")
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
		log.Err(err).Debug("Failed to create client-auth keyfile")
		return errors.WithStack(err)
	}
	keyfile := strings.TrimSpace(cert) + "\n" + strings.TrimSpace(key)

	// Create secrets (in memory)
	keyfileSecret := core.Secret{
		TypeMeta: meta.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: meta.ObjectMeta{
			Name: apSecretName + "-auth",
			Labels: map[string]string{
				labelKeyOriginalDeployment: d.currentObject.GetName(),
			},
		},
		Data: map[string][]byte{
			constants.SecretTLSKeyfile: []byte(keyfile),
		},
		Type: "Opaque",
	}
	tlsCASecret := core.Secret{
		TypeMeta: meta.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: meta.ObjectMeta{
			Name: apSecretName + "-ca",
			Labels: map[string]string{
				labelKeyOriginalDeployment: d.currentObject.GetName(),
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
		log.Err(err).Debug("Failed to encode client-auth keyfile Secret")
		return errors.WithStack(err)
	}
	tlsCAYaml, err := yaml.Marshal(tlsCASecret)
	if err != nil {
		log.Err(err).Debug("Failed to encode TLS CA Secret")
		return errors.WithStack(err)
	}
	allYaml := strings.TrimSpace(string(keyfileYaml)) + "\n---\n" + strings.TrimSpace(string(tlsCAYaml))

	// Create secret containing access package
	secret := &core.Secret{
		ObjectMeta: meta.ObjectMeta{
			Name: apSecretName,
		},
		Data: map[string][]byte{
			constants.SecretAccessPackageYaml: []byte(allYaml),
			constants.SecretCACertificate:     []byte(tlsCACert),
			constants.SecretTLSKeyfile:        []byte(keyfile),
		},
	}
	// Attach secret to owner
	secret.SetOwnerReferences(append(secret.GetOwnerReferences(), d.currentObject.AsOwner()))
	err = globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
		_, err := d.SecretsModInterface().Create(ctxChild, secret, meta.CreateOptions{})
		return err
	})
	if err != nil {
		// Failed to create secret
		log.Err(err).Str("secret-name", apSecretName).Debug("Failed to create access package Secret")
		return errors.WithStack(err)
	}

	// Write sectionLogger entry & create event
	log.Str("secret-name", apSecretName).Info("Created access package Secret")
	d.CreateEvent(k8sutil.NewAccessPackageCreatedEvent(d.currentObject, apSecretName))

	return nil
}
