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
	"fmt"
	"sort"
	"time"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"

	"github.com/arangodb/kube-arangodb/pkg/deployment/features"

	"github.com/arangodb/kube-arangodb/pkg/deployment/client"

	"github.com/rs/zerolog/log"
	core "k8s.io/api/core/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/pod"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/rs/zerolog"
)

func createJWTKeyUpdate(ctx context.Context,
	log zerolog.Logger, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	cachedStatus inspectorInterface.Inspector, context PlanBuilderContext) api.Plan {
	if folder, err := ensureJWTFolderSupport(spec, status); err != nil || !folder {
		return nil
	}

	folder, ok := cachedStatus.Secret(pod.JWTSecretFolder(apiObject.GetName()))
	if !ok {
		log.Error().Msgf("Unable to get JWT folder info")
		return nil
	}

	s, ok := cachedStatus.Secret(spec.Authentication.GetJWTSecretName())
	if !ok {
		log.Info().Msgf("JWT Secret is missing, no rotation will take place")
		return nil
	}

	jwt, ok := s.Data[constants.SecretKeyToken]
	if !ok {
		log.Warn().Msgf("JWT Secret is invalid, no rotation will take place")
		return addJWTPropagatedPlanAction(status)
	}

	jwtSha := util.SHA256(jwt)

	if _, ok := folder.Data[jwtSha]; !ok {
		return addJWTPropagatedPlanAction(status, api.NewAction(api.ActionTypeJWTAdd, api.ServerGroupUnknown, "", "Add JWTRotation key").AddParam(checksum, jwtSha))
	}

	activeKey, ok := folder.Data[pod.ActiveJWTKey]
	if !ok {
		return addJWTPropagatedPlanAction(status, api.NewAction(api.ActionTypeJWTSetActive, api.ServerGroupUnknown, "", "Set active key").AddParam(checksum, jwtSha))
	}

	tokenKey, ok := folder.Data[constants.SecretKeyToken]
	if !ok || util.SHA256(activeKey) != util.SHA256(tokenKey) {
		return addJWTPropagatedPlanAction(status, api.NewAction(api.ActionTypeJWTSetActive, api.ServerGroupUnknown, "", "Set active key and add token field").AddParam(checksum, jwtSha))
	}

	plan, failed := areJWTTokensUpToDate(ctx, log, status, context, folder)
	if len(plan) > 0 {
		return plan
	}

	if failed {
		log.Info().Msgf("JWT Failed on one pod, no rotation will take place")
		return nil
	}

	if util.SHA256(activeKey) != jwtSha {
		return addJWTPropagatedPlanAction(status, api.NewAction(api.ActionTypeJWTSetActive, api.ServerGroupUnknown, "", "Set active key").AddParam(checksum, jwtSha))
	}

	for key := range folder.Data {
		if key == pod.ActiveJWTKey || key == constants.SecretKeyToken {
			continue
		}

		if key == jwtSha {
			continue
		}

		return addJWTPropagatedPlanAction(status, api.NewAction(api.ActionTypeJWTClean, api.ServerGroupUnknown, "", "Remove old key").AddParam(checksum, key))
	}

	return addJWTPropagatedPlanAction(status)
}

func createJWTStatusUpdate(ctx context.Context,
	log zerolog.Logger, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	cachedStatus inspectorInterface.Inspector, context PlanBuilderContext) api.Plan {
	if _, err := ensureJWTFolderSupport(spec, status); err != nil {
		return nil
	}

	if createJWTStatusUpdateRequired(log, apiObject, spec, status, cachedStatus) {
		return addJWTPropagatedPlanAction(status, api.NewAction(api.ActionTypeJWTStatusUpdate, api.ServerGroupUnknown, "", "Update status"))
	}

	return nil
}

func createJWTStatusUpdateRequired(log zerolog.Logger, apiObject k8sutil.APIObject, spec api.DeploymentSpec,
	status api.DeploymentStatus, cachedStatus inspectorInterface.Inspector) bool {
	folder, err := ensureJWTFolderSupport(spec, status)
	if err != nil {
		log.Error().Err(err).Msgf("Action not supported")
		return false
	}

	if !folder {
		if status.Hashes.JWT.Passive != nil {
			return true
		}

		f, ok := cachedStatus.Secret(spec.Authentication.GetJWTSecretName())
		if !ok {
			log.Error().Msgf("Unable to get JWT secret info")
			return false
		}

		key, ok := f.Data[constants.SecretKeyToken]
		if !ok {
			log.Error().Msgf("JWT Token is invalid")
			return false
		}

		keySha := fmt.Sprintf("sha256:%s", util.SHA256(key))

		if status.Hashes.JWT.Active != keySha {
			log.Error().Msgf("JWT Token is invalid")
			return true
		}

		return false
	}

	f, ok := cachedStatus.Secret(pod.JWTSecretFolder(apiObject.GetName()))
	if !ok {
		log.Error().Msgf("Unable to get JWT folder info")
		return false
	}

	activeKeyData, active := f.Data[pod.ActiveJWTKey]
	activeKeyShort := util.SHA256(activeKeyData)
	activeKey := fmt.Sprintf("sha256:%s", activeKeyShort)
	if active {
		if status.Hashes.JWT.Active != activeKey {
			return true
		}
	}

	if len(f.Data) == 0 {
		return status.Hashes.JWT.Passive != nil
	}

	var keys []string

	for key := range f.Data {
		if key == pod.ActiveJWTKey || key == activeKeyShort || key == constants.SecretKeyToken {
			continue
		}

		keys = append(keys, key)
	}

	if len(keys) == 0 {
		return status.Hashes.JWT.Passive != nil
	}

	sort.Strings(keys)
	keys = util.PrefixStringArray(keys, "sha256:")

	return !util.CompareStringArray(keys, status.Hashes.JWT.Passive)
}

func areJWTTokensUpToDate(ctx context.Context, log zerolog.Logger, status api.DeploymentStatus,
	planCtx PlanBuilderContext, folder *core.Secret) (plan api.Plan, failed bool) {
	gCtx, c := context.WithTimeout(ctx, 2*time.Second)
	defer c()

	status.Members.ForeachServerGroup(func(group api.ServerGroup, list api.MemberStatusList) error {
		for _, m := range list {
			nCtx, c := context.WithTimeout(gCtx, 500*time.Millisecond)
			defer c()
			if updateRequired, failedMember := isJWTTokenUpToDate(nCtx, log, status, planCtx, group, m, folder); failedMember {
				failed = true
				continue
			} else if updateRequired {
				plan = append(plan, api.NewAction(api.ActionTypeJWTRefresh, group, m.ID))
				continue
			}
		}

		return nil
	})

	return
}

func isJWTTokenUpToDate(ctx context.Context, log zerolog.Logger, status api.DeploymentStatus, context PlanBuilderContext,
	group api.ServerGroup, m api.MemberStatus, folder *core.Secret) (updateRequired bool, failed bool) {
	if m.Phase != api.MemberPhaseCreated {
		return false, true
	}

	if i, ok := status.Images.GetByImageID(m.ImageID); !ok || !features.JWTRotation().Supported(i.ArangoDBVersion, i.Enterprise) {
		return false, false
	}

	mlog := log.With().Str("group", group.AsRole()).Str("member", m.ID).Logger()

	c, err := context.GetServerClient(ctx, group, m.ID)
	if err != nil {
		mlog.Warn().Err(err).Msg("Unable to get client")
		return false, true
	}

	if updateRequired, err := isMemberJWTTokenInvalid(ctx, client.NewClient(c.Connection()), folder.Data, false); err != nil {
		mlog.Warn().Err(err).Msg("JWT UpToDate Check failed")
		return false, true
	} else if updateRequired {
		return true, false
	}

	return false, false
}

func addJWTPropagatedPlanAction(s api.DeploymentStatus, actions ...api.Action) api.Plan {
	got := len(actions) != 0
	cond := conditionFalse
	if !got {
		cond = conditionTrue
	}

	if s.Hashes.JWT.Propagated == got {
		p := api.Plan{api.NewAction(api.ActionTypeJWTPropagated, api.ServerGroupUnknown, "", "Change propagated flag").AddParam(propagated, cond)}
		return append(p, actions...)
	}

	return actions
}

func isMemberJWTTokenInvalid(ctx context.Context, c client.Client, data map[string][]byte, refresh bool) (bool, error) {
	cmd := c.GetJWT
	if refresh {
		cmd = c.RefreshJWT
	}

	e, err := cmd(ctx)
	if err != nil {
		return false, errors.Wrapf(err, "Unable to fetch JWT tokens")
	}

	if e.Result.Active == nil {
		return false, errors.Wrapf(err, "There is no active JWT Token")
	}

	if jwtActive, ok := data[pod.ActiveJWTKey]; !ok {
		return false, errors.Newf("Missing Active JWT Token in folder")
	} else if util.SHA256(jwtActive) != e.Result.Active.GetSHA().Checksum() {
		log.Info().Str("active", e.Result.Active.GetSHA().Checksum()).Str("expected", util.SHA256(jwtActive)).Msgf("Active key is invalid")
		return true, nil
	}

	if !compareJWTKeys(e.Result.Passive, data) {
		return true, nil
	}

	return false, nil
}

func compareJWTKeys(e client.Entries, keys map[string][]byte) bool {
	for k := range keys {
		if k == pod.ActiveJWTKey || k == constants.SecretKeyToken {
			continue
		}

		if !e.Contains(k) {
			log.Info().Msgf("Missing JWT Key")
			return false
		}
	}

	for _, entry := range e {
		if entry.GetSHA() == "" {
			continue
		}

		if _, ok := keys[entry.GetSHA().Checksum()]; !ok {
			return false
		}
	}

	return true
}
