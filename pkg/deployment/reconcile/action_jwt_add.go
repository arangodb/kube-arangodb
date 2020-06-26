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
	"encoding/base64"

	"github.com/arangodb/kube-arangodb/pkg/util/constants"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/patch"
	"github.com/arangodb/kube-arangodb/pkg/deployment/pod"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/rs/zerolog"
	"k8s.io/apimachinery/pkg/types"
)

func init() {
	registerAction(api.ActionTypeJWTAdd, newJWTAdd)
}

func newJWTAdd(log zerolog.Logger, action api.Action, actionCtx ActionContext) Action {
	a := &jwtAddAction{}

	a.actionImpl = newActionImplDefRef(log, action, actionCtx, defaultTimeout)

	return a
}

type jwtAddAction struct {
	actionImpl

	actionEmptyCheckProgress
}

func (a *jwtAddAction) Start(ctx context.Context) (bool, error) {
	folder, err := ensureJWTFolderSupportFromAction(a.actionCtx)
	if err != nil {
		a.log.Error().Err(err).Msgf("Action not supported")
		return true, nil
	}

	if !folder {
		a.log.Error().Msgf("Action not supported")
		return true, nil
	}

	appendToken, exists := a.action.Params[checksum]
	if !exists {
		a.log.Warn().Msgf("Key %s is missing in action", checksum)
		return true, nil
	}

	s, ok := a.actionCtx.GetCachedStatus().Secret(a.actionCtx.GetSpec().Authentication.GetJWTSecretName())
	if !ok {
		a.log.Error().Msgf("JWT Secret is missing, no rotation will take place")
		return true, nil
	}

	jwt, ok := s.Data[constants.SecretKeyToken]
	if !ok {
		a.log.Error().Msgf("JWT Secret is invalid, no rotation will take place")
		return true, nil
	}

	jwtSha := util.SHA256(jwt)

	if appendToken != jwtSha {
		a.log.Error().Msgf("JWT Secret changed")
		return true, nil
	}

	f, ok := a.actionCtx.GetCachedStatus().Secret(pod.JWTSecretFolder(a.actionCtx.GetName()))
	if !ok {
		a.log.Error().Msgf("Unable to get JWT folder info")
		return true, nil
	}

	if _, ok := f.Data[jwtSha]; ok {
		a.log.Info().Msgf("JWT Already exists")
		return true, nil
	}

	p := patch.NewPatch()
	p.ItemAdd(patch.NewPath("data", jwtSha), base64.StdEncoding.EncodeToString(jwt))

	patch, err := p.Marshal()
	if err != nil {
		a.log.Error().Err(err).Msgf("Unable to encrypt patch")
		return true, nil
	}

	_, err = a.actionCtx.SecretsInterface().Patch(pod.JWTSecretFolder(a.actionCtx.GetName()), types.JSONPatchType, patch)
	if err != nil {
		if !k8sutil.IsInvalid(err) {
			return false, errors.Wrapf(err, "Unable to update secret: %s", pod.JWTSecretFolder(a.actionCtx.GetName()))
		}
	}

	return true, nil
}
