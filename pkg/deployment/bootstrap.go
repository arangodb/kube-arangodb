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

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	driver "github.com/arangodb/go-driver"

	"github.com/arangodb/kube-arangodb/pkg/util/constants"
)

const (
	rootUserName = "root"
)

// EnsureBootstrap executes the bootstrap once as soon as the deployment becomes ready
func (d *Deployment) EnsureBootstrap() error {

	status, version := d.GetStatus()

	if status.Conditions.IsTrue(api.ConditionTypeReady) {
		if status.Conditions.IsTrue(api.ConditionTypeBoostrapCompleted) {
			return nil // Nothing to do, already bootstrapped
		}

		d.deps.Log.Info().Msgf("Bootstrap deployment %s", d.Name())
		err := d.runBootstrap()
		if err != nil {
			status.Conditions.Update(api.ConditionTypeBoostrapCompleted, true, "Bootstrap failed", err.Error())
		} else {
			status.Conditions.Update(api.ConditionTypeBoostrapCompleted, true, "Bootstrap successful", "The bootstrap process has been completed")
		}

		if err = d.UpdateStatus(status, version); err != nil {
			return maskAny(err)
		}

		d.deps.Log.Info().Msgf("Bootstrap completed for %s", d.Name())
	}

	return nil
}

func (d *Deployment) ensureRootUserPassword() (string, error) {

	spec := d.GetSpec()
	secrets := d.GetKubeCli().CoreV1().Secrets(d.Namespace())
	if auth, err := secrets.Get(spec.GetRootUserAccessSecretName(), metav1.GetOptions{}); k8sutil.IsNotFound(err) {
		// Create new one
		tokenData := make([]byte, 32)
		rand.Read(tokenData)
		token := hex.EncodeToString(tokenData)
		owner := d.GetAPIObject().AsOwner()

		if err := k8sutil.CreateBasicAuthSecret(secrets, spec.GetRootUserAccessSecretName(), rootUserName, token, &owner); err != nil {
			return "", err
		}

		return token, nil
	} else if err == nil {
		user, ok := auth.Data[constants.SecretUsername]
		if ok && string(user) == rootUserName {
			pass, ok := auth.Data[constants.SecretPassword]
			if ok {
				return string(pass), nil
			}
		}
		return "", fmt.Errorf("invalid secret format")
	} else {
		return "", err
	}
}

func (d *Deployment) runBootstrap() error {

	// execute the boostrap code
	// make sure that the bootstrap code is idempotent
	client, err := d.clientCache.GetDatabase(nil)
	if err != nil {
		return maskAny(err)
	}

	password, err := d.ensureRootUserPassword()
	if err != nil {
		return maskAny(err)
	}

	// Obtain the root user
	root, err := client.User(nil, rootUserName)
	if err != nil {
		return maskAny(err)
	}

	err = root.Update(nil, driver.UserOptions{
		Password: password,
	})
	if err != nil {
		return maskAny(err)
	}

	return nil
}
