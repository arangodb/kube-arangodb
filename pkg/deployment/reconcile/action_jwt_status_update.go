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
	"fmt"
	"sort"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	"github.com/arangodb/kube-arangodb/pkg/deployment/pod"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/strings"
)

const (
	checksum   = "checksum"
	propagated = "propagated"

	conditionTrue  = "True"
	conditionFalse = "False"
)

func ensureJWTFolderSupportFromAction(actionCtx ActionContext) (bool, error) {
	return ensureJWTFolderSupport(actionCtx.GetSpec(), actionCtx.GetStatus())
}

func ensureJWTFolderSupport(spec api.DeploymentSpec, status api.DeploymentStatus) (bool, error) {
	if !spec.IsAuthenticated() {
		return false, errors.Newf("Authentication is disabled")
	}

	if image := status.CurrentImage; image == nil {
		return false, errors.Newf("Missing image info")
	} else {
		if !features.JWTRotation().Supported(image.ArangoDBVersion, image.Enterprise) {
			return false, nil
		}
	}
	return true, nil
}

func newJWTStatusUpdateAction(action api.Action, actionCtx ActionContext) Action {
	a := &actionJWTStatusUpdate{}

	a.actionImpl = newActionImplDefRef(action, actionCtx)

	return a
}

type actionJWTStatusUpdate struct {
	actionImpl

	actionEmptyCheckProgress
}

func (a *actionJWTStatusUpdate) Start(ctx context.Context) (bool, error) {
	folder, err := ensureJWTFolderSupportFromAction(a.actionCtx)
	if err != nil {
		a.log.Err(err).Error("Action not supported")
		return true, nil
	}

	if !folder {
		f, ok := a.actionCtx.ACS().CurrentClusterCache().Secret().V1().GetSimple(a.actionCtx.GetSpec().Authentication.GetJWTSecretName())
		if !ok {
			a.log.Error("Unable to get JWT secret info")
			return true, nil
		}

		key, ok := f.Data[constants.SecretKeyToken]
		if !ok {
			a.log.Error("JWT Token is invalid")
			return true, nil
		}

		keySha := fmt.Sprintf("sha256:%s", util.SHA256(key))

		if err = a.actionCtx.WithStatusUpdate(ctx, func(s *api.DeploymentStatus) bool {
			if s.Hashes.JWT.Passive != nil {
				s.Hashes.JWT.Passive = nil
				return true
			}

			if s.Hashes.JWT.Active != keySha {
				s.Hashes.JWT.Active = keySha
				return true
			}

			return false
		}); err != nil {
			return false, err
		}

		return true, nil
	}

	f, ok := a.actionCtx.ACS().CurrentClusterCache().Secret().V1().GetSimple(pod.JWTSecretFolder(a.actionCtx.GetName()))
	if !ok {
		a.log.Error("Unable to get JWT folder info")
		return true, nil
	}

	if err = a.actionCtx.WithStatusUpdate(ctx, func(s *api.DeploymentStatus) (update bool) {
		activeKeyData, active := f.Data[pod.ActiveJWTKey]
		activeKeyShort := util.SHA256(activeKeyData)
		activeKey := fmt.Sprintf("sha256:%s", activeKeyShort)
		if active {
			if s.Hashes.JWT.Active != activeKey {
				s.Hashes.JWT.Active = activeKey
				update = true
			}
		}

		if len(f.Data) == 0 {
			if s.Hashes.JWT.Passive != nil {
				s.Hashes.JWT.Passive = nil
				update = true
			}
		}

		var keys []string

		for key := range f.Data {
			if key == pod.ActiveJWTKey || key == activeKeyShort || key == constants.SecretKeyToken {
				continue
			}

			keys = append(keys, key)
		}

		if len(keys) == 0 {
			if s.Hashes.JWT.Passive != nil {
				s.Hashes.JWT.Passive = nil
				update = true
			}
		}

		sort.Strings(keys)
		keys = strings.PrefixStringArray(keys, "sha256:")

		if !strings.CompareStringArray(keys, s.Hashes.JWT.Passive) {
			s.Hashes.JWT.Passive = keys
			update = true
		}

		return
	}); err != nil {
		return false, err
	}

	return true, nil
}
