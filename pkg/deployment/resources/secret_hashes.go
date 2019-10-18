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
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"github.com/arangodb/go-driver"

	v1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

// ValidateSecretHashes checks the hash of used secrets
// against the stored ones.
// If a hash is different, the deployment is marked
// with a SecretChangedCondition and the operator will not
// touch it until this is resolved.
func (r *Resources) ValidateSecretHashes() error {
	// validate performs a secret hash comparison for a single secret.
	// Return true if all is good, false when the SecretChanged condition
	// must be set.
	validate := func(secretName string,
		getExpectedHash func() string,
		setExpectedHash func(string) error,
		actionHashChanged func(Context, *v1.Secret) error) (bool, error) {

		log := r.log.With().Str("secret-name", secretName).Logger()
		expectedHash := getExpectedHash()
		secret, hash, err := r.getSecretHash(secretName)
		if expectedHash == "" {
			// No hash set yet, try to fill it
			if k8sutil.IsNotFound(err) {
				// Secret does not (yet) exists, do nothing
				return true, nil
			}
			if err != nil {
				log.Warn().Err(err).Msg("Failed to get secret")
				return true, nil // Since we do not yet have a hash, we let this go with only a warning.
			}
			// Hash fetched succesfully, store it
			if err := setExpectedHash(hash); err != nil {
				log.Debug().Msg("Failed to save secret hash")
				return true, maskAny(err)
			}
			return true, nil
		}
		// Hash is set, it must match the current hash
		if err != nil {
			// Fetching error failed for other reason.
			log.Debug().Err(err).Msg("Failed to fetch secret hash")
			// This is not good, return false so SecretsChanged condition will be set.
			return false, nil
		}
		if hash != expectedHash {
			// Oops, hash has changed
			log.Debug().
				Str("expected-hash", expectedHash).
				Str("new-hash", hash).
				Msg("Secret has changed.")
			if actionHashChanged != nil {
				if err := actionHashChanged(r.context, secret); err != nil {
					log.Debug().Msgf("failed to change secret. hash-changed-action returned error: %v", err)
					return true, nil
				}

				if err := setExpectedHash(hash); err != nil {
					log.Debug().Msg("Failed to change secret hash")
					return true, maskAny(err)
				}
				return true, nil
			}
			// This is not good, return false so SecretsChanged condition will be set.
			return false, nil
		}
		// All good
		return true, nil
	}

	spec := r.context.GetSpec()
	log := r.log
	var badSecretNames []string
	status, lastVersion := r.context.GetStatus()
	getHashes := func() *api.SecretHashes {
		if status.SecretHashes == nil {
			status.SecretHashes = api.NewEmptySecretHashes()
		}
		if status.SecretHashes.Users == nil {
			status.SecretHashes.Users = make(map[string]string)
		}
		return status.SecretHashes
	}
	updateHashes := func(updater func(*api.SecretHashes)) error {
		if status.SecretHashes == nil {
			status.SecretHashes = api.NewEmptySecretHashes()
		}
		if status.SecretHashes.Users == nil {
			status.SecretHashes.Users = make(map[string]string)
		}
		updater(status.SecretHashes)
		if err := r.context.UpdateStatus(status, lastVersion); err != nil {
			return maskAny(err)
		}
		// Reload status
		status, lastVersion = r.context.GetStatus()
		return nil
	}

	if spec.IsAuthenticated() {
		secretName := spec.Authentication.GetJWTSecretName()
		getExpectedHash := func() string { return getHashes().AuthJWT }
		setExpectedHash := func(h string) error {
			return maskAny(updateHashes(func(dst *api.SecretHashes) { dst.AuthJWT = h }))
		}
		if hashOK, err := validate(secretName, getExpectedHash, setExpectedHash, nil); err != nil {
			return maskAny(err)
		} else if !hashOK {
			badSecretNames = append(badSecretNames, secretName)
		}
	}
	if spec.RocksDB.IsEncrypted() {
		secretName := spec.RocksDB.Encryption.GetKeySecretName()
		getExpectedHash := func() string { return getHashes().RocksDBEncryptionKey }
		setExpectedHash := func(h string) error {
			return maskAny(updateHashes(func(dst *api.SecretHashes) { dst.RocksDBEncryptionKey = h }))
		}
		if hashOK, err := validate(secretName, getExpectedHash, setExpectedHash, nil); err != nil {
			return maskAny(err)
		} else if !hashOK {
			badSecretNames = append(badSecretNames, secretName)
		}
	}
	if spec.IsSecure() {
		secretName := spec.TLS.GetCASecretName()
		getExpectedHash := func() string { return getHashes().TLSCA }
		setExpectedHash := func(h string) error {
			return maskAny(updateHashes(func(dst *api.SecretHashes) { dst.TLSCA = h }))
		}
		if hashOK, err := validate(secretName, getExpectedHash, setExpectedHash, nil); err != nil {
			return maskAny(err)
		} else if !hashOK {
			badSecretNames = append(badSecretNames, secretName)
		}
	}
	if spec.Sync.IsEnabled() {
		secretName := spec.Sync.TLS.GetCASecretName()
		getExpectedHash := func() string { return getHashes().SyncTLSCA }
		setExpectedHash := func(h string) error {
			return maskAny(updateHashes(func(dst *api.SecretHashes) { dst.SyncTLSCA = h }))
		}
		if hashOK, err := validate(secretName, getExpectedHash, setExpectedHash, nil); err != nil {
			return maskAny(err)
		} else if !hashOK {
			badSecretNames = append(badSecretNames, secretName)
		}
	}

	for username, secretName := range spec.Bootstrap.PasswordSecretNames {
		if secretName.IsNone() || secretName.IsAuto() {
			continue
		}

		_, err := r.context.GetKubeCli().CoreV1().Secrets(r.context.GetNamespace()).Get(string(secretName), metav1.GetOptions{})
		if k8sutil.IsNotFound(err) {
			// do nothing when secret was deleted
			continue
		}

		getExpectedHash := func() string {
			if v, ok := getHashes().Users[username]; ok {
				return v
			}
			return ""
		}
		setExpectedHash := func(h string) error {
			return maskAny(updateHashes(func(dst *api.SecretHashes) {
				dst.Users[username] = h
			}))
		}

		// If password changes it should not be set that deployment in 'SecretsChanged' state
		validate(string(secretName), getExpectedHash, setExpectedHash, changeUserPassword)
	}

	if len(badSecretNames) > 0 {
		// We have invalid hashes, set the SecretsChanged condition
		if status.Conditions.Update(api.ConditionTypeSecretsChanged, true,
			"Secrets have changed", fmt.Sprintf("Found %d changed secrets", len(badSecretNames))) {
			log.Warn().Msgf("Found %d changed secrets. Settings SecretsChanged condition", len(badSecretNames))
			if err := r.context.UpdateStatus(status, lastVersion); err != nil {
				log.Error().Err(err).Msg("Failed to save SecretsChanged condition")
				return maskAny(err)
			}
			// Add an event about this
			r.context.CreateEvent(k8sutil.NewSecretsChangedEvent(badSecretNames, r.context.GetAPIObject()))
		}
	} else {
		// All good, we van remove the SecretsChanged condition
		if status.Conditions.Remove(api.ConditionTypeSecretsChanged) {
			log.Info().Msg("Resetting SecretsChanged condition")
			if err := r.context.UpdateStatus(status, lastVersion); err != nil {
				log.Error().Err(err).Msg("Failed to save SecretsChanged condition")
				return maskAny(err)
			}
			// Add an event about this
			r.context.CreateEvent(k8sutil.NewSecretsRestoredEvent(r.context.GetAPIObject()))
		}
	}

	return nil
}

func changeUserPassword(c Context, secret *v1.Secret) error {
	username, password, err := k8sutil.GetSecretAuthCredentials(secret)
	if err != nil {
		return nil
	}

	ctx := context.Background()
	client, err := c.GetDatabaseClient(ctx)
	if err != nil {
		return maskAny(err)
	}

	user, err := client.User(ctx, username)
	if err != nil {
		if driver.IsNotFound(err) {
			options := &driver.UserOptions{
				Password: password,
				Active:   new(bool),
			}
			*options.Active = true

			_, err = client.CreateUser(ctx, username, options)
			return maskAny(err)
		}
		return err
	}

	err = user.Update(ctx, driver.UserOptions{
		Password: password,
	})

	return maskAny(err)
}

// getSecretHash fetches a secret with given name and returns a hash over its value.
func (r *Resources) getSecretHash(secretName string) (*v1.Secret, string, error) {
	kubecli := r.context.GetKubeCli()
	ns := r.context.GetNamespace()
	s, err := kubecli.CoreV1().Secrets(ns).Get(secretName, metav1.GetOptions{})
	if err != nil {
		return nil, "", maskAny(err)
	}
	// Create hash of value
	rows := make([]string, 0, len(s.Data))
	for k, v := range s.Data {
		rows = append(rows, k+"="+hex.EncodeToString(v))
	}
	// Sort so we're not detecting order differences
	sort.Strings(rows)
	data := strings.Join(rows, "\n")
	rawHash := sha256.Sum256([]byte(data))
	hash := fmt.Sprintf("%0x", rawHash)
	return s, hash, nil
}
