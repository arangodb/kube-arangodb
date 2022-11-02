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
	"encoding/base64"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/patch"
	"github.com/arangodb/kube-arangodb/pkg/deployment/pod"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
)

func newJWTAddAction(action api.Action, actionCtx ActionContext) Action {
	a := &actionJWTAdd{}

	a.actionImpl = newActionImplDefRef(action, actionCtx)

	return a
}

type actionJWTAdd struct {
	actionImpl

	actionEmptyCheckProgress
}

func (a *actionJWTAdd) Start(ctx context.Context) (bool, error) {
	folder, err := ensureJWTFolderSupportFromAction(a.actionCtx)
	if err != nil {
		a.log.Err(err).Error("Action not supported")
		return true, nil
	}

	if !folder {
		a.log.Error("Action not supported")
		return true, nil
	}

	appendToken, exists := a.action.Params[checksum]
	if !exists {
		a.log.Warn("Key %s is missing in action", checksum)
		return true, nil
	}

	s, ok := a.actionCtx.ACS().CurrentClusterCache().Secret().V1().GetSimple(a.actionCtx.GetSpec().Authentication.GetJWTSecretName())
	if !ok {
		a.log.Error("JWT Secret is missing, no rotation will take place")
		return true, nil
	}

	jwt, ok := s.Data[constants.SecretKeyToken]
	if !ok {
		a.log.Error("JWT Secret is invalid, no rotation will take place")
		return true, nil
	}

	jwtSha := util.SHA256(jwt)

	if appendToken != jwtSha {
		a.log.Error("JWT Secret changed")
		return true, nil
	}

	f, ok := a.actionCtx.ACS().CurrentClusterCache().Secret().V1().GetSimple(pod.JWTSecretFolder(a.actionCtx.GetName()))
	if !ok {
		a.log.Error("Unable to get JWT folder info")
		return true, nil
	}

	if _, ok := f.Data[jwtSha]; ok {
		a.log.Info("JWT Already exists")
		return true, nil
	}

	p := patch.NewPatch()
	p.ItemAdd(patch.NewPath("data", jwtSha), base64.StdEncoding.EncodeToString(jwt))

	patch, err := p.Marshal()
	if err != nil {
		a.log.Err(err).Error("Unable to encrypt patch")
		return true, nil
	}

	err = globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
		_, err := a.actionCtx.ACS().CurrentClusterCache().SecretsModInterface().V1().Patch(ctxChild, pod.JWTSecretFolder(a.actionCtx.GetName()), types.JSONPatchType, patch, meta.PatchOptions{})
		return err
	})
	if err != nil {
		if !kerrors.IsInvalid(err) {
			return false, errors.Wrapf(err, "Unable to update secret: %s", pod.JWTSecretFolder(a.actionCtx.GetName()))
		}
	}

	return true, nil
}
