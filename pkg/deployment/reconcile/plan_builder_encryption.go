//
// DISCLAIMER
//
// Copyright 2020-2021 ArangoDB GmbH, Cologne, Germany
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
// Author Adam Janikowski
// Author Tomasz Mielech
//

package reconcile

import (
	"context"

	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod"

	core "k8s.io/api/core/v1"

	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"

	"github.com/arangodb/kube-arangodb/pkg/util"

	"github.com/arangodb/kube-arangodb/pkg/deployment/client"
	"github.com/arangodb/kube-arangodb/pkg/deployment/pod"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/rs/zerolog"
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

func createEncryptionKeyStatusPropagatedFieldUpdate(ctx context.Context,
	log zerolog.Logger, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	cachedStatus inspectorInterface.Inspector, context PlanBuilderContext, w WithPlanBuilder, builders ...planBuilder) api.Plan {
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
			api.NewAction(api.ActionTypeEncryptionKeyPropagated, api.ServerGroupUnknown, "", "Change propagated flag to false").AddParam(propagated, conditionFalse),
		}, plan...)
	}

	return plan
}

func createEncryptionKey(ctx context.Context,
	log zerolog.Logger, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	cachedStatus inspectorInterface.Inspector, context PlanBuilderContext) api.Plan {
	if skipEncryptionPlan(spec, status) {
		return nil
	}

	secret, exists := cachedStatus.Secret(spec.RocksDB.Encryption.GetKeySecretName())
	if !exists {
		return nil
	}

	name, _, err := pod.GetEncryptionKeyFromSecret(secret)
	if err != nil {
		log.Error().Err(err).Msgf("Unable to fetch encryption key")
		return nil
	}

	if !exists {
		return nil
	}

	keyfolder, exists := cachedStatus.Secret(pod.GetEncryptionFolderSecretName(context.GetName()))
	if !exists {
		log.Error().Msgf("Encryption key folder does not exist")
		return nil
	}

	if len(keyfolder.Data) == 0 {
		keyfolder.Data = map[string][]byte{}
	}

	if status.Hashes.Encryption.Propagated {
		_, ok := keyfolder.Data[name]
		if !ok {
			return api.Plan{api.NewAction(api.ActionTypeEncryptionKeyAdd, api.ServerGroupUnknown, "")}
		}
	}

	plan, failed := areEncryptionKeysUpToDate(ctx, log, spec, status, context, keyfolder)
	if !plan.IsEmpty() {
		return plan
	}

	if !failed && !status.Hashes.Encryption.Propagated {
		return api.Plan{
			api.NewAction(api.ActionTypeEncryptionKeyPropagated, api.ServerGroupUnknown, "", "Change propagated flag to true").AddParam(propagated, conditionTrue),
		}
	}

	return api.Plan{}
}

func createEncryptionKeyStatusUpdate(ctx context.Context,
	log zerolog.Logger, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	cachedStatus inspectorInterface.Inspector, context PlanBuilderContext) api.Plan {
	if skipEncryptionPlan(spec, status) {
		return nil
	}

	if createEncryptionKeyStatusUpdateRequired(log, spec, status, cachedStatus, context) {
		return api.Plan{api.NewAction(api.ActionTypeEncryptionKeyStatusUpdate, api.ServerGroupUnknown, "")}
	}

	return nil

}

func createEncryptionKeyStatusUpdateRequired(log zerolog.Logger, spec api.DeploymentSpec, status api.DeploymentStatus,
	cachedStatus inspectorInterface.Inspector, context PlanBuilderContext) bool {
	if skipEncryptionPlan(spec, status) {
		return false
	}

	keyfolder, exists := cachedStatus.Secret(pod.GetEncryptionFolderSecretName(context.GetName()))
	if !exists {
		log.Error().Msgf("Encryption key folder does not exist")
		return false
	}

	keyHashes := secretKeysToListWithPrefix(keyfolder)

	return !util.CompareStringArray(keyHashes, status.Hashes.Encryption.Keys)
}

func createEncryptionKeyCleanPlan(ctx context.Context,
	log zerolog.Logger, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	cachedStatus inspectorInterface.Inspector, context PlanBuilderContext) api.Plan {
	if skipEncryptionPlan(spec, status) {
		return nil
	}

	keyfolder, exists := cachedStatus.Secret(pod.GetEncryptionFolderSecretName(context.GetName()))
	if !exists {
		log.Error().Msgf("Encryption key folder does not exist")
		return nil
	}

	if !status.Hashes.Encryption.Propagated {
		return nil
	}

	var plan api.Plan

	if len(keyfolder.Data) <= 1 {
		return nil
	}

	secret, exists := cachedStatus.Secret(spec.RocksDB.Encryption.GetKeySecretName())
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
		log.Err(err).Msgf("Key from encryption is not in keyfolder - do nothing")
		return nil
	}

	for key := range keyfolder.Data {
		if key != name {
			plan = append(plan, api.NewAction(api.ActionTypeEncryptionKeyRemove, api.ServerGroupUnknown, "").AddParam("key", key))
		}
	}

	if !plan.IsEmpty() {
		return plan
	}

	return api.Plan{}
}

func areEncryptionKeysUpToDate(ctx context.Context, log zerolog.Logger, spec api.DeploymentSpec,
	status api.DeploymentStatus, context PlanBuilderContext, folder *core.Secret) (plan api.Plan, failed bool) {

	status.Members.ForeachServerGroup(func(group api.ServerGroup, list api.MemberStatusList) error {
		if !pod.GroupEncryptionSupported(spec.Mode.Get(), group) {
			return nil
		}

		for _, m := range list {
			if updateRequired, failedMember := isEncryptionKeyUpToDate(ctx, log, status, context, group, m, folder); failedMember {
				failed = true
				continue
			} else if updateRequired {
				plan = append(plan, api.NewAction(api.ActionTypeEncryptionKeyRefresh, group, m.ID))
				continue
			}
		}

		return nil
	})

	return
}

func isEncryptionKeyUpToDate(ctx context.Context,
	log zerolog.Logger, status api.DeploymentStatus,
	planCtx PlanBuilderContext,
	group api.ServerGroup, m api.MemberStatus,
	folder *core.Secret) (updateRequired bool, failed bool) {
	if m.Phase != api.MemberPhaseCreated {
		return false, true
	}

	if i, ok := status.Images.GetByImageID(m.ImageID); !ok || !features.EncryptionRotation().Supported(i.ArangoDBVersion, i.Enterprise) {
		return false, false
	}

	mlog := log.With().Str("group", group.AsRole()).Str("member", m.ID).Logger()

	ctxChild, cancel := context.WithTimeout(ctx, arangod.GetRequestTimeout())
	defer cancel()
	c, err := planCtx.GetServerClient(ctxChild, group, m.ID)
	if err != nil {
		mlog.Warn().Err(err).Msg("Unable to get client")
		return false, true
	}

	client := client.NewClient(c.Connection())

	ctxChild, cancel = context.WithTimeout(ctx, arangod.GetRequestTimeout())
	defer cancel()
	e, err := client.GetEncryption(ctxChild)
	if err != nil {
		mlog.Error().Err(err).Msgf("Unable to fetch encryption keys")
		return false, true
	}

	if !e.Result.KeysPresent(folder.Data) {
		mlog.Info().Msgf("Refresh of encryption keys required")
		return true, false
	}

	return false, false
}
