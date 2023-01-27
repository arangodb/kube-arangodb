//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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

	core "k8s.io/api/core/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/actions"
	"github.com/arangodb/kube-arangodb/pkg/deployment/client"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	"github.com/arangodb/kube-arangodb/pkg/deployment/pod"
	"github.com/arangodb/kube-arangodb/pkg/deployment/reconcile/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/strings"
)

func skipEncryptionPlan(spec api.DeploymentSpec, status api.DeploymentStatus) bool {
	if !spec.RocksDB.IsEncrypted() {
		return true
	}

	if i := status.CurrentImage; i == nil || !features.EncryptionRotation().Supported(i.ArangoDBVersion, i.Enterprise) {
		return true
	}

	return false
}

func (r *Reconciler) createEncryptionKeyStatusPropagatedFieldUpdate(ctx context.Context, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	context PlanBuilderContext, w WithPlanBuilder, builders ...planBuilder) api.Plan {
	if skipEncryptionPlan(spec, status) {
		return nil
	}

	var plan api.Plan

	for _, builder := range builders {
		if !plan.IsEmpty() {
			continue
		}

		if p := w.Apply(builder); !p.IsEmpty() {
			plan = append(plan, p...)
		}
	}

	if plan.IsEmpty() {
		return nil
	}

	if len(plan) == 1 && plan[0].Type == api.ActionTypeEncryptionKeyPropagated {
		return plan
	}

	if status.Hashes.Encryption.Propagated {
		plan = append(api.Plan{
			actions.NewClusterAction(api.ActionTypeEncryptionKeyPropagated, "Change propagated flag to false").AddParam(propagated, conditionFalse),
		}, plan...)
	}

	return plan
}

func (r *Reconciler) createEncryptionKey(ctx context.Context, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	context PlanBuilderContext) api.Plan {
	if skipEncryptionPlan(spec, status) {
		return nil
	}

	secret, exists := context.ACS().CurrentClusterCache().Secret().V1().GetSimple(spec.RocksDB.Encryption.GetKeySecretName())
	if !exists {
		return nil
	}

	name, _, err := pod.GetEncryptionKeyFromSecret(secret)
	if err != nil {
		r.log.Err(err).Error("Unable to fetch encryption key")
		return nil
	}

	if !exists {
		return nil
	}

	keyfolder, exists := context.ACS().CurrentClusterCache().Secret().V1().GetSimple(pod.GetEncryptionFolderSecretName(context.GetName()))
	if !exists {
		r.log.Error("Encryption key folder does not exist")
		return nil
	}

	if len(keyfolder.Data) == 0 {
		keyfolder.Data = map[string][]byte{}
	}

	if status.Hashes.Encryption.Propagated {
		_, ok := keyfolder.Data[name]
		if !ok {
			return api.Plan{actions.NewClusterAction(api.ActionTypeEncryptionKeyAdd)}
		}
	}

	plan, failed := r.areEncryptionKeysUpToDate(ctx, spec, status, context, keyfolder)
	if !plan.IsEmpty() {
		return plan
	}

	if !failed && !status.Hashes.Encryption.Propagated {
		return api.Plan{
			actions.NewClusterAction(api.ActionTypeEncryptionKeyPropagated, "Change propagated flag to true").AddParam(propagated, conditionTrue),
		}
	}

	return api.Plan{}
}

func (r *Reconciler) createEncryptionKeyStatusUpdate(ctx context.Context, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	context PlanBuilderContext) api.Plan {
	if skipEncryptionPlan(spec, status) {
		return nil
	}

	if r.createEncryptionKeyStatusUpdateRequired(spec, status, context) {
		return api.Plan{actions.NewClusterAction(api.ActionTypeEncryptionKeyStatusUpdate)}
	}

	return nil

}

func (r *Reconciler) createEncryptionKeyStatusUpdateRequired(spec api.DeploymentSpec, status api.DeploymentStatus,
	context PlanBuilderContext) bool {
	if skipEncryptionPlan(spec, status) {
		return false
	}

	keyfolder, exists := context.ACS().CurrentClusterCache().Secret().V1().GetSimple(pod.GetEncryptionFolderSecretName(context.GetName()))
	if !exists {
		r.log.Error("Encryption key folder does not exist")
		return false
	}

	keyHashes := secretKeysToListWithPrefix(keyfolder)

	return !strings.CompareStringArray(keyHashes, status.Hashes.Encryption.Keys)
}

func (r *Reconciler) createEncryptionKeyCleanPlan(ctx context.Context, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	context PlanBuilderContext) api.Plan {
	if skipEncryptionPlan(spec, status) {
		return nil
	}

	keyfolder, exists := context.ACS().CurrentClusterCache().Secret().V1().GetSimple(pod.GetEncryptionFolderSecretName(context.GetName()))
	if !exists {
		r.log.Error("Encryption key folder does not exist")
		return nil
	}

	if !status.Hashes.Encryption.Propagated {
		return nil
	}

	var plan api.Plan

	if len(keyfolder.Data) <= 1 {
		return nil
	}

	secret, exists := context.ACS().CurrentClusterCache().Secret().V1().GetSimple(spec.RocksDB.Encryption.GetKeySecretName())
	if !exists {
		return nil
	}

	name, _, err := pod.GetEncryptionKeyFromSecret(secret)
	if err != nil {
		return nil
	}

	if !exists {
		return nil
	}

	if _, ok := keyfolder.Data[name]; !ok {
		r.log.Err(err).Error("Key from encryption is not in keyfolder - do nothing")
		return nil
	}

	for key := range keyfolder.Data {
		if key != name {
			plan = append(plan, actions.NewClusterAction(api.ActionTypeEncryptionKeyRemove).AddParam("key", key))
		}
	}

	if !plan.IsEmpty() {
		return plan
	}

	return api.Plan{}
}

func (r *Reconciler) areEncryptionKeysUpToDate(ctx context.Context, spec api.DeploymentSpec,
	status api.DeploymentStatus, context PlanBuilderContext, folder *core.Secret) (plan api.Plan, failed bool) {

	for _, group := range api.AllServerGroups {
		if !pod.GroupEncryptionSupported(spec.Mode.Get(), group) {
			continue
		}
		for _, m := range status.Members.MembersOfGroup(group) {
			if updateRequired, failedMember := r.isEncryptionKeyUpToDate(ctx, status, context, group, m, folder); failedMember {
				failed = true
				continue
			} else if updateRequired {
				plan = append(plan, actions.NewAction(api.ActionTypeEncryptionKeyRefresh, group, shared.WithPredefinedMember(m.ID)))
				continue
			}
		}
	}

	return
}

func (r *Reconciler) isEncryptionKeyUpToDate(ctx context.Context, status api.DeploymentStatus,
	planCtx PlanBuilderContext,
	group api.ServerGroup, m api.MemberStatus,
	folder *core.Secret) (updateRequired bool, failed bool) {
	if m.Phase != api.MemberPhaseCreated {
		return false, true
	}

	if i, ok := status.Images.GetByImageID(m.ImageID); !ok || !features.EncryptionRotation().Supported(i.ArangoDBVersion, i.Enterprise) {
		return false, false
	}

	log := r.log.Str("group", group.AsRole()).Str("member", m.ID)
	c, err := planCtx.GetMembersState().GetMemberClient(m.ID)
	if err != nil {
		log.Err(err).Warn("Unable to get client")
		return false, true
	}

	client := client.NewClient(c.Connection(), log)

	ctxChild, cancel := globals.GetGlobalTimeouts().ArangoD().WithTimeout(ctx)
	defer cancel()
	e, err := client.GetEncryption(ctxChild)
	if err != nil {
		log.Err(err).Error("Unable to fetch encryption keys")
		return false, true
	}

	if !e.Result.KeysPresent(folder.Data) {
		log.Info("Refresh of encryption keys required")
		return true, false
	}

	return false, false
}
