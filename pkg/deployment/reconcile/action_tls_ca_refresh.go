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
	"encoding/base64"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/patch"
	"github.com/arangodb/kube-arangodb/pkg/deployment/resources"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
)

// newRefreshTLSCAAction creates a new Action that implements the given
// planned RefreshTLSCA action.
func newRefreshTLSCAAction(action api.Action, actionCtx ActionContext) Action {
	a := &actionRefreshTLSCA{}

	a.actionImpl = newBaseActionImpl(action, actionCtx, &a.newMemberID)

	return a
}

// actionRefreshTLSCA implements an RefreshTLSCAAction.
type actionRefreshTLSCA struct {
	// actionImpl implement timeout and member id functions
	actionImpl

	// actionEmptyCheckProgress implement check progress with empty implementation
	actionEmptyCheckProgress

	newMemberID string
}

// Start performs the start of the action.
// Returns true if the action is completely finished, false in case
// the start time needs to be recorded and a ready condition needs to be checked.
func (a *actionRefreshTLSCA) Start(ctx context.Context) (bool, error) {
	caFolder, exists := a.actionCtx.ACS().CurrentClusterCache().Secret().V1().GetSimple(resources.GetCASecretName(a.actionCtx.GetAPIObject()))
	if !exists {
		a.log.Warn("Secret %s is missing", resources.GetCASecretName(a.actionCtx.GetAPIObject()))
		return true, nil
	}

	caLocalData := buildInternalCA(caFolder)

	if string(caFolder.Data[resources.CACertName]) != caLocalData {
		p := patch.NewPatch()
		p.ItemAdd(patch.NewPath("data", resources.CACertName), base64.StdEncoding.EncodeToString([]byte(caLocalData)))

		patch, err := p.Marshal()
		if err != nil {
			a.log.Err(err).Error("Unable to encrypt patch")
			return true, nil
		}

		err = globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
			_, err := a.actionCtx.ACS().CurrentClusterCache().SecretsModInterface().V1().Patch(ctxChild, resources.GetCASecretName(a.actionCtx.GetAPIObject()), types.JSONPatchType, patch, meta.PatchOptions{})
			return err
		})
		if err != nil {
			if !kerrors.IsInvalid(err) {
				return false, errors.Wrapf(err, "Unable to update secret: %s", string(patch))
			}
		}
	}

	return true, nil
}
