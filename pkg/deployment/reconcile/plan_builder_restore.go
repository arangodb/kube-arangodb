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

	backupv1 "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/pod"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/rs/zerolog"
)

func createRestorePlan(ctx context.Context, log zerolog.Logger, spec api.DeploymentSpec, status api.DeploymentStatus, builderCtx PlanBuilderContext) api.Plan {
	if spec.RestoreFrom == nil && status.Restore != nil {
		return api.Plan{
			api.NewAction(api.ActionTypeBackupRestoreClean, api.ServerGroupUnknown, ""),
		}
	}

	if spec.RestoreFrom != nil && status.Restore == nil {
		backup, err := builderCtx.GetBackup(spec.GetRestoreFrom())
		if err != nil {
			log.Warn().Err(err).Msg("Backup not found")
			return nil
		}

		if p := createRestorePlanEncryption(ctx, log, spec, status, builderCtx, backup); !p.IsEmpty() {
			return p
		}

		if backup.Status.Backup == nil {
			log.Warn().Msg("Backup not yet ready")
			return nil
		}

		return api.Plan{
			api.NewAction(api.ActionTypeBackupRestore, api.ServerGroupUnknown, ""),
		}
	}

	return nil
}

func createRestorePlanEncryption(ctx context.Context, log zerolog.Logger, spec api.DeploymentSpec, status api.DeploymentStatus, builderCtx PlanBuilderContext, backup *backupv1.ArangoBackup) api.Plan {
	if backup.Spec.EncryptionSecret != nil {
		if !spec.RocksDB.IsEncrypted() {
			return nil
		}

		if i := status.CurrentImage; i == nil || !i.Enterprise || i.ArangoDBVersion.CompareTo("3.7.0") < 0 {
			return nil
		}

		secret := *backup.Spec.EncryptionSecret

		// Additional logic to do restore with encryption key
		keyfolder, err := builderCtx.SecretsInterface().Get(pod.GetKeyfolderSecretName(builderCtx.GetName()), meta.GetOptions{})
		if err != nil {
			log.Err(err).Msgf("Unable to fetch encryption folder")
			return nil
		}

		if len(keyfolder.Data) <= 1 {
			return nil
		}
		name, _, err := pod.GetEncryptionKey(builderCtx.SecretsInterface(), secret)
		if err != nil {
			log.Err(err).Msgf("Unable to fetch encryption key")
			return nil
		}

		if _, ok := keyfolder.Data[name]; !ok {
			log.Err(err).Msgf("Key from encryption is not in keyfolder - first install this secret")
			return nil
		}

		return api.Plan{
			api.NewAction(api.ActionTypeEncryptionKeyAdd, api.ServerGroupUnknown, "").AddParam("secret", secret),
		}
	}

	return nil
}
