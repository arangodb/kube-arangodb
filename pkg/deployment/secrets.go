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
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/k8s-operator/pkg/apis/arangodb/v1alpha"
	"github.com/arangodb/k8s-operator/pkg/util/k8sutil"
)

// createSecrets creates all secrets needed to run the given deployment
func (d *Deployment) createSecrets(apiObject *api.ArangoDeployment) error {
	if apiObject.Spec.IsAuthenticated() {
		if err := d.ensureJWTSecret(apiObject.Spec.Authentication.JWTSecretName); err != nil {
			return maskAny(err)
		}
	}
	return nil
}

// ensureJWTSecret checks if a secret with given name exists in the namespace
// of the deployment. If not, it will add such a secret with a random
// JWT token.
func (d *Deployment) ensureJWTSecret(secretName string) error {
	kubecli := d.deps.KubeCli
	ns := d.apiObject.GetNamespace()
	if _, err := kubecli.CoreV1().Secrets(ns).Get(secretName, metav1.GetOptions{}); k8sutil.IsNotFound(err) {
		// Secret not found, create it
		// Create token
		tokenData := make([]byte, 32)
		rand.Read(tokenData)
		token := hex.EncodeToString(tokenData)

		// Create secret
		secret := &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: secretName,
			},
			Data: map[string][]byte{
				"token": []byte(token),
			},
		}
		// Attach secret to deployment
		secret.SetOwnerReferences(append(secret.GetOwnerReferences(), d.apiObject.AsOwner()))
		if _, err := kubecli.CoreV1().Secrets(ns).Create(secret); k8sutil.IsAlreadyExists(err) {
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
func (d *Deployment) getJWTSecret(apiObject *api.ArangoDeployment) (string, error) {
	if !apiObject.Spec.IsAuthenticated() {
		return "", nil
	}
	kubecli := d.deps.KubeCli
	secretName := apiObject.Spec.Authentication.JWTSecretName
	s, err := k8sutil.GetJWTSecret(kubecli, secretName, apiObject.GetNamespace())
	if err != nil {
		d.deps.Log.Debug().Err(err).Str("secret-name", secretName).Msg("Failed to get JWT secret")
		return "", maskAny(err)
	}
	return s, nil
}

// getSyncJWTSecret loads the JWT secret used for syncmasters from a Secret configured in apiObject.Spec.Sync.Authentication.JWTSecretName.
func (d *Deployment) getSyncJWTSecret(apiObject *api.ArangoDeployment) (string, error) {
	kubecli := d.deps.KubeCli
	secretName := apiObject.Spec.Sync.Authentication.JWTSecretName
	s, err := k8sutil.GetJWTSecret(kubecli, secretName, apiObject.GetNamespace())
	if err != nil {
		d.deps.Log.Debug().Err(err).Str("secret-name", secretName).Msg("Failed to get sync JWT secret")
		return "", maskAny(err)
	}
	return s, nil
}

// getSyncMonitoringToken loads the token secret used for monitoring sync masters & workers.
func (d *Deployment) getSyncMonitoringToken(apiObject *api.ArangoDeployment) (string, error) {
	kubecli := d.deps.KubeCli
	secretName := apiObject.Spec.Sync.Monitoring.TokenSecretName
	s, err := kubecli.CoreV1().Secrets(apiObject.GetNamespace()).Get(secretName, metav1.GetOptions{})
	if err != nil {
		d.deps.Log.Debug().Err(err).Str("secret-name", secretName).Msg("Failed to get monitoring token secret")
	}
	// Take the first data
	for _, v := range s.Data {
		return string(v), nil
	}
	return "", maskAny(fmt.Errorf("No data found in secret '%s'", secretName))
}
