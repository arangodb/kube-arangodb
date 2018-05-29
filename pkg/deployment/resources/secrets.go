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
	"crypto/rand"
	"encoding/hex"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

// EnsureSecrets creates all secrets needed to run the given deployment
func (r *Resources) EnsureSecrets() error {
	spec := r.context.GetSpec()
	if spec.IsAuthenticated() {
		if err := r.ensureTokenSecret(spec.Authentication.GetJWTSecretName()); err != nil {
			return maskAny(err)
		}
	}
	if spec.IsSecure() {
		if err := r.ensureTLSCACertificateSecret(spec.TLS); err != nil {
			return maskAny(err)
		}
	}
	if spec.Sync.IsEnabled() {
		if err := r.ensureTokenSecret(spec.Sync.Authentication.GetJWTSecretName()); err != nil {
			return maskAny(err)
		}
		if err := r.ensureTokenSecret(spec.Sync.Monitoring.GetTokenSecretName()); err != nil {
			return maskAny(err)
		}
		if err := r.ensureTLSCACertificateSecret(spec.Sync.TLS); err != nil {
			return maskAny(err)
		}
		if err := r.ensureClientAuthCACertificateSecret(spec.Sync.Authentication); err != nil {
			return maskAny(err)
		}
	}
	return nil
}

// ensureTokenSecret checks if a secret with given name exists in the namespace
// of the deployment. If not, it will add such a secret with a random
// token.
func (r *Resources) ensureTokenSecret(secretName string) error {
	kubecli := r.context.GetKubeCli()
	ns := r.context.GetNamespace()
	if _, err := kubecli.CoreV1().Secrets(ns).Get(secretName, metav1.GetOptions{}); k8sutil.IsNotFound(err) {
		// Secret not found, create it
		// Create token
		tokenData := make([]byte, 32)
		rand.Read(tokenData)
		token := hex.EncodeToString(tokenData)

		// Create secret
		owner := r.context.GetAPIObject().AsOwner()
		if err := k8sutil.CreateTokenSecret(kubecli.CoreV1(), secretName, ns, token, &owner); k8sutil.IsAlreadyExists(err) {
			// Secret added while we tried it also
			return nil
		} else if err != nil {
			// Failed to create secret
			return maskAny(err)
		}
	} else if err != nil {
		// Failed to get secret for other reasons
		return maskAny(err)
	}
	return nil
}

// ensureTLSCACertificateSecret checks if a secret with given name exists in the namespace
// of the deployment. If not, it will add such a secret with a generated CA certificate.
func (r *Resources) ensureTLSCACertificateSecret(spec api.TLSSpec) error {
	kubecli := r.context.GetKubeCli()
	ns := r.context.GetNamespace()
	if _, err := kubecli.CoreV1().Secrets(ns).Get(spec.GetCASecretName(), metav1.GetOptions{}); k8sutil.IsNotFound(err) {
		// Secret not found, create it
		apiObject := r.context.GetAPIObject()
		owner := apiObject.AsOwner()
		deploymentName := apiObject.GetName()
		if err := createTLSCACertificate(r.log, kubecli.CoreV1(), spec, deploymentName, ns, &owner); k8sutil.IsAlreadyExists(err) {
			// Secret added while we tried it also
			return nil
		} else if err != nil {
			// Failed to create secret
			return maskAny(err)
		}
	} else if err != nil {
		// Failed to get secret for other reasons
		return maskAny(err)
	}
	return nil
}

// ensureClientAuthCACertificateSecret checks if a secret with given name exists in the namespace
// of the deployment. If not, it will add such a secret with a generated CA certificate.
func (r *Resources) ensureClientAuthCACertificateSecret(spec api.SyncAuthenticationSpec) error {
	kubecli := r.context.GetKubeCli()
	ns := r.context.GetNamespace()
	if _, err := kubecli.CoreV1().Secrets(ns).Get(spec.GetClientCASecretName(), metav1.GetOptions{}); k8sutil.IsNotFound(err) {
		// Secret not found, create it
		apiObject := r.context.GetAPIObject()
		owner := apiObject.AsOwner()
		deploymentName := apiObject.GetName()
		if err := createClientAuthCACertificate(r.log, kubecli.CoreV1(), spec, deploymentName, ns, &owner); k8sutil.IsAlreadyExists(err) {
			// Secret added while we tried it also
			return nil
		} else if err != nil {
			// Failed to create secret
			return maskAny(err)
		}
	} else if err != nil {
		// Failed to get secret for other reasons
		return maskAny(err)
	}
	return nil
}

// getJWTSecret loads the JWT secret from a Secret configured in apiObject.Spec.Authentication.JWTSecretName.
func (r *Resources) getJWTSecret(spec api.DeploymentSpec) (string, error) {
	if !spec.IsAuthenticated() {
		return "", nil
	}
	kubecli := r.context.GetKubeCli()
	ns := r.context.GetNamespace()
	secretName := spec.Authentication.GetJWTSecretName()
	s, err := k8sutil.GetTokenSecret(kubecli.CoreV1(), secretName, ns)
	if err != nil {
		r.log.Debug().Err(err).Str("secret-name", secretName).Msg("Failed to get JWT secret")
		return "", maskAny(err)
	}
	return s, nil
}

// getSyncJWTSecret loads the JWT secret used for syncmasters from a Secret configured in apiObject.Spec.Sync.Authentication.JWTSecretName.
func (r *Resources) getSyncJWTSecret(spec api.DeploymentSpec) (string, error) {
	kubecli := r.context.GetKubeCli()
	ns := r.context.GetNamespace()
	secretName := spec.Sync.Authentication.GetJWTSecretName()
	s, err := k8sutil.GetTokenSecret(kubecli.CoreV1(), secretName, ns)
	if err != nil {
		r.log.Debug().Err(err).Str("secret-name", secretName).Msg("Failed to get sync JWT secret")
		return "", maskAny(err)
	}
	return s, nil
}

// getSyncMonitoringToken loads the token secret used for monitoring sync masters & workers.
func (r *Resources) getSyncMonitoringToken(spec api.DeploymentSpec) (string, error) {
	kubecli := r.context.GetKubeCli()
	ns := r.context.GetNamespace()
	secretName := spec.Sync.Monitoring.GetTokenSecretName()
	s, err := k8sutil.GetTokenSecret(kubecli.CoreV1(), secretName, ns)
	if err != nil {
		r.log.Debug().Err(err).Str("secret-name", secretName).Msg("Failed to get sync monitoring secret")
		return "", maskAny(err)
	}
	return s, nil
}
