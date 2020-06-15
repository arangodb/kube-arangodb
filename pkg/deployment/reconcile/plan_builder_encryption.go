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
// Author Adam Janikowski
//

package reconcile

import (
	"context"

	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"

	"github.com/arangodb/kube-arangodb/pkg/deployment/resources/inspector"

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

	if i := status.CurrentImage; i == nil || !i.Enterprise || i.ArangoDBVersion.CompareTo("3.7.0") < 0 {
		return true
	}

	return false
}

func createEncryptionKey(ctx context.Context,
	log zerolog.Logger, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	cachedStatus inspector.Inspector, context PlanBuilderContext) api.Plan {
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

	keyfolder, exists := cachedStatus.Secret(pod.GetKeyfolderSecretName(context.GetName()))
	if !exists {
		log.Error().Msgf("Encryption key folder does not exist")
		return nil
	}

	if len(keyfolder.Data) == 0 {
		keyfolder.Data = map[string][]byte{}
	}

	_, ok := keyfolder.Data[name]
	if !ok {
		return api.Plan{api.NewAction(api.ActionTypeEncryptionKeyAdd, api.ServerGroupUnknown, "")}
	}

	var plan api.Plan
	status.Members.ForeachServerGroup(func(group api.ServerGroup, members api.MemberStatusList) error {
		if !pod.GroupEncryptionSupported(spec.Mode.Get(), group) {
			return nil
		}

		glog := log.With().Str("group", group.AsRole())

		for _, m := range members {
			if m.Phase != api.MemberPhaseCreated {
				// Only make changes when phase is created
				continue
			}

			if m.ArangoVersion.CompareTo("3.7.0") < 0 {
				continue
			}

			mlog := glog.Str("member", m.ID).Logger()

			c, err := context.GetServerClient(ctx, group, m.ID)
			if err != nil {
				mlog.Warn().Err(err).Msg("Unable to get client")
				continue
			}

			client := client.NewClient(c.Connection())

			e, err := client.GetEncryption(ctx)
			if err != nil {
				mlog.Error().Err(err).Msgf("Unable to fetch encryption keys")
				continue
			}

			if !e.Result.KeysPresent(keyfolder.Data) {
				plan = append(plan, api.NewAction(api.ActionTypeEncryptionKeyRefresh, group, m.ID))
				mlog.Info().Msgf("Refresh of encryption keys required")
				continue
			}
		}

		return nil
	})

	if !plan.IsEmpty() {
		return plan
	}

	return api.Plan{}
}

func createEncryptionKeyStatusUpdate(ctx context.Context,
	log zerolog.Logger, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	cachedStatus inspector.Inspector, context PlanBuilderContext) api.Plan {
	if skipEncryptionPlan(spec, status) {
		return nil
	}

	if createEncryptionKeyStatusUpdateRequired(ctx, log, apiObject, spec, status, cachedStatus, context) {
		return api.Plan{api.NewAction(api.ActionTypeEncryptionKeyStatusUpdate, api.ServerGroupUnknown, "")}
	}

	return nil

}

func createEncryptionKeyStatusUpdateRequired(ctx context.Context,
	log zerolog.Logger, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	cachedStatus inspector.Inspector, context PlanBuilderContext) bool {
	if skipEncryptionPlan(spec, status) {
		return false
	}

	keyfolder, exists := cachedStatus.Secret(pod.GetKeyfolderSecretName(context.GetName()))
	if !exists {
		log.Error().Msgf("Encryption key folder does not exist")
		return false
	}

	keyHashes := secretKeysToListWithPrefix("sha256:", keyfolder)

	if !util.CompareStringArray(keyHashes, status.Hashes.Encryption) {
		return true
	}

	return false
}

func cleanEncryptionKey(ctx context.Context,
	log zerolog.Logger, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	cachedStatus inspector.Inspector, context PlanBuilderContext) api.Plan {
	if skipEncryptionPlan(spec, status) {
		return nil
	}

	keyfolder, exists := cachedStatus.Secret(pod.GetKeyfolderSecretName(context.GetName()))
	if !exists {
		log.Error().Msgf("Encryption key folder does not exist")
		return nil
	}

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

	var plan api.Plan

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
