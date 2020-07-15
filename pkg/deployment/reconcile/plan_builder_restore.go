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
// Author Tomasz Mielech <tomasz@arangodb.com>
//

package reconcile

import (
	"context"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"

	"github.com/arangodb/kube-arangodb/pkg/deployment/resources/inspector"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"

	backupv1 "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/pod"
	"github.com/rs/zerolog"
)

const secretActionParam = "secret"

func createRestorePlan(ctx context.Context,
	log zerolog.Logger, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	cachedStatus inspector.Inspector, context PlanBuilderContext) api.Plan {
	if spec.RestoreFrom == nil && status.Restore != nil {
		return api.Plan{
			api.NewAction(api.ActionTypeBackupRestoreClean, api.ServerGroupUnknown, ""),
		}
	}

	if spec.RestoreFrom != nil && status.Restore == nil {
		backup, err := context.GetBackup(spec.GetRestoreFrom())
		if err != nil {
			log.Warn().Err(err).Msg("Backup not found")
			return nil
		}

		if backup.Status.Backup == nil {
			log.Warn().Msg("Backup not yet ready")
			return nil
		}

		if spec.RocksDB.IsEncrypted() {
			if ok, p := createRestorePlanEncryption(ctx, log, spec, status, context, backup); !ok {
				return nil
			} else if !p.IsEmpty() {
				return p
			}

			if !status.Hashes.Encryption.Propagated {
				log.Warn().Msg("Backup not able to be restored in non propagated state")
				return nil
			}
		}

		return api.Plan{
			api.NewAction(api.ActionTypeBackupRestore, api.ServerGroupUnknown, ""),
		}
	}

	return nil
}

func createRestorePlanEncryption(ctx context.Context, log zerolog.Logger, spec api.DeploymentSpec, status api.DeploymentStatus, builderCtx PlanBuilderContext, backup *backupv1.ArangoBackup) (bool, api.Plan) {
	if spec.RestoreEncryptionSecret != nil {
		if !spec.RocksDB.IsEncrypted() {
			return true, nil
		}

		if i := status.CurrentImage; i == nil || !features.EncryptionRotation().Supported(i.ArangoDBVersion, i.Enterprise) {
			                        return nil
					}

		if !status.Hashes.Encryption.Propagated {
			return false, nil
		}

		secret := *spec.RestoreEncryptionSecret

		// Additional logic to do restore with encryption key
		name, _, exists, err := pod.GetEncryptionKey(builderCtx.SecretsInterface(), secret)
		if err != nil {
			log.Err(err).Msgf("Unable to fetch encryption key")
			return false, nil
		}

		if !exists {
			log.Error().Msgf("Unable to fetch encryption key - key is empty or missing")
			return false, nil
		}

		if !status.Hashes.Encryption.Keys.ContainsSHA256(name) {
			return true, api.Plan{
				api.NewAction(api.ActionTypeEncryptionKeyPropagated, api.ServerGroupUnknown, "").AddParam(propagated, conditionFalse),
				api.NewAction(api.ActionTypeEncryptionKeyAdd, api.ServerGroupUnknown, "").AddParam(secretActionParam, secret),
			}
		}

		return true, nil
	}

	return true, nil
}
