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
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

// ValidateSecretHashes checks the hash of used secrets
// against the stored ones.
// If a hash is different, the deployment is marked
// with a SecretChangedCondition and the operator will no
// touch it until this is resolved.
func (r *Resources) ValidateSecretHashes() error {
	// validate performs a secret hash comparison for a single secret.
	// Return true if all is good, false when the SecretChanged condition
	// must be set.
	validate := func(secretName string, expectedHashRef *string, status *api.DeploymentStatus) (bool, error) {
		log := r.log.With().Str("secret-name", secretName).Logger()
		expectedHash := *expectedHashRef
		hash, err := r.getSecretHash(secretName)
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
			*expectedHashRef = hash
			if r.context.UpdateStatus(*status); err != nil {
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
			log.Error().Msg("Secret has changed. You must revert it to the original value!")
			// This is not good, return false so SecretsChanged condition will be set.
			return false, nil
		}
		// All good
		return true, nil
	}

	spec := r.context.GetSpec()
	log := r.log
	var badSecretNames []string
	status := r.context.GetStatus()
	if status.SecretHashes == nil {
		status.SecretHashes = &api.SecretHashes{}
	}
	hashes := status.SecretHashes
	if spec.IsAuthenticated() {
		secretName := spec.Authentication.GetJWTSecretName()
		if hashOK, err := validate(secretName, &hashes.AuthJWT, &status); err != nil {
			return maskAny(err)
		} else if !hashOK {
			badSecretNames = append(badSecretNames, secretName)
		}
	}
	if spec.RocksDB.IsEncrypted() {
		secretName := spec.RocksDB.Encryption.GetKeySecretName()
		if hashOK, err := validate(secretName, &hashes.RocksDBEncryptionKey, &status); err != nil {
			return maskAny(err)
		} else if !hashOK {
			badSecretNames = append(badSecretNames, secretName)
		}
	}
	if spec.IsSecure() {
		secretName := spec.TLS.GetCASecretName()
		if hashOK, err := validate(secretName, &hashes.TLSCA, &status); err != nil {
			return maskAny(err)
		} else if !hashOK {
			badSecretNames = append(badSecretNames, secretName)
		}
	}
	if spec.Sync.IsEnabled() {
		secretName := spec.Sync.TLS.GetCASecretName()
		if hashOK, err := validate(secretName, &hashes.SyncTLSCA, &status); err != nil {
			return maskAny(err)
		} else if !hashOK {
			badSecretNames = append(badSecretNames, secretName)
		}
	}

	if len(badSecretNames) > 0 {
		// We have invalid hashes, set the SecretsChanged condition
		if status.Conditions.Update(api.ConditionTypeSecretsChanged, true,
			"Secrets have changed", fmt.Sprintf("Found %d changed secrets", len(badSecretNames))) {
			log.Warn().Msgf("Found %d changed secrets. Settings SecretsChanged condition", len(badSecretNames))
			if err := r.context.UpdateStatus(status); err != nil {
				log.Error().Err(err).Msg("Failed to save SecretsChanged condition")
				return maskAny(err)
			}
			// Add an event about this
			r.context.CreateEvent(k8sutil.NewSecretsChangedEvent(badSecretNames, r.context.GetAPIObject()))
		}
	} else {
		// All good, we van remove the SecretsChanged condition
		if status.Conditions.Remove(api.ConditionTypeSecretsChanged) {
			log.Warn().Msg("Resetting SecretsChanged condition")
			if err := r.context.UpdateStatus(status); err != nil {
				log.Error().Err(err).Msg("Failed to save SecretsChanged condition")
				return maskAny(err)
			}
			// Add an event about this
			r.context.CreateEvent(k8sutil.NewSecretsRestoredEvent(r.context.GetAPIObject()))
		}
	}

	return nil
}

// getSecretHash fetches a secret with given name and returns a hash over its value.
func (r *Resources) getSecretHash(secretName string) (string, error) {
	kubecli := r.context.GetKubeCli()
	ns := r.context.GetNamespace()
	s, err := kubecli.CoreV1().Secrets(ns).Get(secretName, metav1.GetOptions{})
	if err != nil {
		return "", maskAny(err)
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
	return hash, nil
}
