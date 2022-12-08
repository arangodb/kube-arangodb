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

package resources

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	core "k8s.io/api/core/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	"github.com/arangodb/kube-arangodb/pkg/deployment/pod"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
)

// ValidateSecretHashes checks the hash of used secrets
// against the stored ones.
// If a hash is different, the deployment is marked
// with a SecretChangedCondition and the operator will not
// touch it until this is resolved.
func (r *Resources) ValidateSecretHashes(ctx context.Context, cachedStatus inspectorInterface.Inspector) error {
	// validate performs a secret hash comparison for a single secret.
	// Return true if all is good, false when the SecretChanged condition
	// must be set.
	log := r.log.Str("section", "secret-hashes")

	validate := func(secretName string,
		getExpectedHash func() string,
		setExpectedHash func(string) error,
		actionHashChanged func(Context, *core.Secret) error) (bool, error) {

		log := log.Str("secret-name", secretName)
		expectedHash := getExpectedHash()
		secret, hash, exists := r.getSecretHash(cachedStatus, secretName)
		if expectedHash == "" {
			// No hash set yet, try to fill it
			if !exists {
				// Secret does not (yet) exists, do nothing
				return true, nil
			}
			// Hash fetched succesfully, store it
			if err := setExpectedHash(hash); err != nil {
				log.Debug("Failed to save secret hash")
				return true, errors.WithStack(err)
			}
			return true, nil
		}
		// Hash is set, it must match the current hash
		if !exists {
			// Fetching error failed for other reason.
			log.Debug("Secret does not exist")
			// This is not good, return false so SecretsChanged condition will be set.
			return false, nil
		}
		if hash != expectedHash {
			// Oops, hash has changed
			log.Str("expected-hash", expectedHash).
				Str("new-hash", hash).
				Debug("Secret has changed.")
			if actionHashChanged != nil {
				if err := actionHashChanged(r.context, secret); err != nil {
					log.Debug("failed to change secret. hash-changed-action returned error: %v", err)
					return true, nil
				}

				if err := setExpectedHash(hash); err != nil {
					log.Debug("Failed to change secret hash")
					return true, errors.WithStack(err)
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
	deploymentName := r.context.GetAPIObject().GetName()
	var badSecretNames []string
	status := r.context.GetStatus()
	image := status.CurrentImage
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
		if err := r.context.UpdateStatus(ctx, status); err != nil {
			return errors.WithStack(err)
		}
		// Reload status
		status = r.context.GetStatus()
		return nil
	}

	if spec.IsAuthenticated() {
		if image == nil || !features.JWTRotation().Supported(image.ArangoDBVersion, image.Enterprise) {
			secretName := spec.Authentication.GetJWTSecretName()
			getExpectedHash := func() string { return getHashes().AuthJWT }
			setExpectedHash := func(h string) error {
				return errors.WithStack(updateHashes(func(dst *api.SecretHashes) { dst.AuthJWT = h }))
			}
			if hashOK, err := validate(secretName, getExpectedHash, setExpectedHash, nil); err != nil {
				return errors.WithStack(err)
			} else if !hashOK {
				badSecretNames = append(badSecretNames, secretName)
			}
		} else {
			if _, exists := cachedStatus.Secret().V1().GetSimple(pod.JWTSecretFolder(deploymentName)); !exists {
				secretName := spec.Authentication.GetJWTSecretName()
				getExpectedHash := func() string { return getHashes().AuthJWT }
				setExpectedHash := func(h string) error {
					return errors.WithStack(updateHashes(func(dst *api.SecretHashes) { dst.AuthJWT = h }))
				}
				if hashOK, err := validate(secretName, getExpectedHash, setExpectedHash, nil); err != nil {
					return errors.WithStack(err)
				} else if !hashOK {
					badSecretNames = append(badSecretNames, secretName)
				}
			}
		}
	}
	if spec.RocksDB.IsEncrypted() {
		if image == nil || !features.EncryptionRotation().Supported(image.ArangoDBVersion, image.Enterprise) {
			secretName := spec.RocksDB.Encryption.GetKeySecretName()
			getExpectedHash := func() string { return getHashes().RocksDBEncryptionKey }
			setExpectedHash := func(h string) error {
				return errors.WithStack(updateHashes(func(dst *api.SecretHashes) { dst.RocksDBEncryptionKey = h }))
			}
			if hashOK, err := validate(secretName, getExpectedHash, setExpectedHash, nil); err != nil {
				return errors.WithStack(err)
			} else if !hashOK {
				badSecretNames = append(badSecretNames, secretName)
			}
		} else {
			if _, exists := cachedStatus.Secret().V1().GetSimple(pod.GetEncryptionFolderSecretName(deploymentName)); !exists {
				secretName := spec.RocksDB.Encryption.GetKeySecretName()
				getExpectedHash := func() string { return getHashes().RocksDBEncryptionKey }
				setExpectedHash := func(h string) error {
					return errors.WithStack(updateHashes(func(dst *api.SecretHashes) { dst.RocksDBEncryptionKey = h }))
				}
				if hashOK, err := validate(secretName, getExpectedHash, setExpectedHash, nil); err != nil {
					return errors.WithStack(err)
				} else if !hashOK {
					badSecretNames = append(badSecretNames, secretName)
				}
			}
		}
	}
	if r.context.IsSyncEnabled() {
		secretName := spec.Sync.TLS.GetCASecretName()
		getExpectedHash := func() string { return getHashes().SyncTLSCA }
		setExpectedHash := func(h string) error {
			return errors.WithStack(updateHashes(func(dst *api.SecretHashes) { dst.SyncTLSCA = h }))
		}
		if hashOK, err := validate(secretName, getExpectedHash, setExpectedHash, nil); err != nil {
			return errors.WithStack(err)
		} else if !hashOK {
			badSecretNames = append(badSecretNames, secretName)
		}
	}

	if len(badSecretNames) > 0 {
		// We have invalid hashes, set the SecretsChanged condition
		if status.Conditions.Update(api.ConditionTypeSecretsChanged, true,
			"Secrets have changed", fmt.Sprintf("Found %d changed secrets", len(badSecretNames))) {
			log.Warn("Found %d changed secrets. Settings SecretsChanged condition", len(badSecretNames))
			if err := r.context.UpdateStatus(ctx, status); err != nil {
				log.Err(err).Error("Failed to save SecretsChanged condition")
				return errors.WithStack(err)
			}
			// Add an event about this
			r.context.CreateEvent(k8sutil.NewSecretsChangedEvent(badSecretNames, r.context.GetAPIObject()))
		}
	} else {
		// All good, we van remove the SecretsChanged condition
		if status.Conditions.Remove(api.ConditionTypeSecretsChanged) {
			log.Info("Resetting SecretsChanged condition")
			if err := r.context.UpdateStatus(ctx, status); err != nil {
				log.Err(err).Error("Failed to save SecretsChanged condition")
				return errors.WithStack(err)
			}
			// Add an event about this
			r.context.CreateEvent(k8sutil.NewSecretsRestoredEvent(r.context.GetAPIObject()))
		}
	}

	return nil
}

// getSecretHash fetches a secret with given name and returns a hash over its value.
func (r *Resources) getSecretHash(cachedStatus inspectorInterface.Inspector, secretName string) (*core.Secret, string, bool) {
	s, exists := cachedStatus.Secret().V1().GetSimple(secretName)
	if !exists {
		return nil, "", false
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
	return s, hash, true
}
