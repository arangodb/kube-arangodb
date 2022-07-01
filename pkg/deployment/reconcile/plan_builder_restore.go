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

package reconcile

import (
	"context"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/actions"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	"github.com/arangodb/kube-arangodb/pkg/deployment/pod"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

const secretActionParam = "secret"

func (r *Reconciler) createRestorePlan(ctx context.Context, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	context PlanBuilderContext) api.Plan {
	if spec.RestoreFrom == nil && status.Restore != nil {
		return api.Plan{
			actions.NewClusterAction(api.ActionTypeBackupRestoreClean),
		}
	}

	if spec.RestoreFrom != nil && status.Restore == nil {
		backup, err := context.GetBackup(ctx, spec.GetRestoreFrom())
		if err != nil {
			r.planLogger.Err(err).Warn("Backup not found")
			return nil
		}

		if backup.Status.Backup == nil {
			r.planLogger.Warn("Backup not yet ready")
			return nil
		}

		if spec.RocksDB.IsEncrypted() {
			if ok, p := r.createRestorePlanEncryption(ctx, spec, status, context); !ok {
				return nil
			} else if !p.IsEmpty() {
				return p
			}

			if i := status.CurrentImage; i != nil && features.EncryptionRotation().Supported(i.ArangoDBVersion, i.Enterprise) {
				if !status.Hashes.Encryption.Propagated {
					r.planLogger.Warn("Backup not able to be restored in non propagated state")
					return nil
				}
			}
		}

		return restorePlan(spec)
	}

	return nil
}

func restorePlan(spec api.DeploymentSpec) api.Plan {
	p := api.Plan{
		actions.NewClusterAction(api.ActionTypeBackupRestore),
	}

	switch spec.Mode.Get() {
	case api.DeploymentModeActiveFailover:
		p = withMaintenance(p...)
	}

	return p
}

func (r *Reconciler) createRestorePlanEncryption(ctx context.Context, spec api.DeploymentSpec, status api.DeploymentStatus, builderCtx PlanBuilderContext) (bool, api.Plan) {

	if spec.RestoreEncryptionSecret != nil {
		if !spec.RocksDB.IsEncrypted() {
			return true, nil
		}

		if i := status.CurrentImage; i == nil || !features.EncryptionRotation().Supported(i.ArangoDBVersion, i.Enterprise) {
			return false, nil
		}

		if !status.Hashes.Encryption.Propagated {
			return false, nil
		}

		secret := *spec.RestoreEncryptionSecret

		// Additional logic to do restore with encryption key
		name, _, exists, err := pod.GetEncryptionKey(ctx, builderCtx.ACS().CurrentClusterCache().Secret().V1().Read(), secret)
		if err != nil {
			r.planLogger.Err(err).Error("Unable to fetch encryption key")
			return false, nil
		}

		if !exists {
			r.planLogger.Error("Unable to fetch encryption key - key is empty or missing")
			return false, nil
		}

		if !status.Hashes.Encryption.Keys.ContainsSHA256(name) {
			return true, api.Plan{
				actions.NewClusterAction(api.ActionTypeEncryptionKeyPropagated).AddParam(propagated, conditionFalse),
				actions.NewClusterAction(api.ActionTypeEncryptionKeyAdd).AddParam(secretActionParam, secret),
			}
		}

		return true, nil
	}

	return true, nil
}
