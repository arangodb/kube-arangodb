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
	"crypto/rand"
	"encoding/hex"
	"fmt"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/go-driver"
)

// EnsureBootstrap executes the bootstrap once as soon as the deployment becomes ready
func (d *Deployment) EnsureBootstrap() error {

	status, version := d.GetStatus()

	if status.Conditions.IsTrue(api.ConditionTypeReady) {
		if _, hasBootstrap := status.Conditions.Get(api.ConditionTypeBootstrapCompleted); !hasBootstrap {
			return nil // The cluster was not initialised with ConditionTypeBoostrapCompleted == false
		}

		if status.Conditions.IsTrue(api.ConditionTypeBootstrapCompleted) {
			return nil // Nothing to do, already bootstrapped
		}

		d.deps.Log.Info().Msgf("Bootstrap deployment %s", d.Name())
		err := d.runBootstrap()
		if err != nil {
			status.Conditions.Update(api.ConditionTypeBootstrapCompleted, true, "Bootstrap failed", err.Error())
			status.Conditions.Update(api.ConditionTypeBootstrapSucceded, false, "Bootstrap failed", err.Error())
		} else {
			status.Conditions.Update(api.ConditionTypeBootstrapCompleted, true, "Bootstrap successful", "The bootstrap process has been completed successfully")
			status.Conditions.Update(api.ConditionTypeBootstrapSucceded, true, "Bootstrap successful", "The bootstrap process has been completed successfully")
		}

		if err = d.UpdateStatus(status, version); err != nil {
			return maskAny(err)
		}

		d.deps.Log.Info().Msgf("Bootstrap completed for %s", d.Name())
	}

	return nil
}

// ensureRootUserPassword ensures the root user secret and returns the password specified or generated
func (d *Deployment) ensureUserPasswordSecret(secrets k8sutil.SecretInterface, username, secretName string) (string, error) {

	if auth, err := secrets.Get(secretName, metav1.GetOptions{}); k8sutil.IsNotFound(err) {
		// Create new one
		tokenData := make([]byte, 32)
		if _, err = rand.Read(tokenData); err != nil {
			return "", err
		}
		token := hex.EncodeToString(tokenData)
		owner := d.GetAPIObject().AsOwner()

		if err := k8sutil.CreateBasicAuthSecret(secrets, secretName, username, token, &owner); err != nil {
			return "", err
		}

		return token, nil
	} else if err == nil {
		user, pass, err := k8sutil.GetSecretAuthCredentials(auth)
		if err == nil && user == username {
			return pass, nil
		}
		return "", fmt.Errorf("invalid secret format in secret %s", secretName)
	} else {
		return "", err
	}
}

// bootstrapUserPassword loads the password for the given user and updates the password stored in the database
func (d *Deployment) bootstrapUserPassword(client driver.Client, secrets k8sutil.SecretInterface, username, secretname string) error {

	d.deps.Log.Debug().Msgf("Bootstrapping user %s, secret %s", username, secretname)

	password, err := d.ensureUserPasswordSecret(secrets, username, secretname)
	if err != nil {
		return maskAny(err)
	}

	// Obtain the user
	if user, err := client.User(context.TODO(), username); driver.IsNotFound(err) {
		_, err := client.CreateUser(context.TODO(), username, &driver.UserOptions{Password: password})
		return maskAny(err)
	} else if err == nil {
		return maskAny(user.Update(context.TODO(), driver.UserOptions{
			Password: password,
		}))
	} else {
		return err
	}
}

// runBootstrap is run for a deployment once
func (d *Deployment) runBootstrap() error {

	// execute the bootstrap code
	// make sure that the bootstrap code is idempotent
	ctx := context.Background()
	client, err := d.clientCache.GetDatabase(ctx)
	if err != nil {
		return maskAny(err)
	}

	spec := d.GetSpec()
	secrets := d.GetKubeCli().CoreV1().Secrets(d.Namespace())

	for user, secret := range spec.Bootstrap.PasswordSecretNames {
		if secret.IsNone() {
			continue
		}
		if err := d.bootstrapUserPassword(client, secrets, user, string(secret)); err != nil {
			return maskAny(err)
		}
	}

	return nil
}
