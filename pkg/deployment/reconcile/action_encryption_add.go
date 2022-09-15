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
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	"github.com/arangodb/kube-arangodb/pkg/deployment/patch"
	"github.com/arangodb/kube-arangodb/pkg/deployment/pod"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
)

func ensureEncryptionSupport(actionCtx ActionContext) error {
	if !actionCtx.GetSpec().RocksDB.IsEncrypted() {
		return errors.Newf("Encryption is disabled")
	}

	if image, ok := actionCtx.GetCurrentImageInfo(); !ok {
		return errors.Newf("Missing image info")
	} else {
		if !features.EncryptionRotation().Supported(image.ArangoDBVersion, image.Enterprise) {
			return errors.Newf("Supported only in Enterprise Edition 3.7.0+")
		}
	}
	return nil
}

func newEncryptionKeyAddAction(action api.Action, actionCtx ActionContext) Action {
	a := &actionEncryptionKeyAdd{}

	a.actionImpl = newActionImplDefRef(action, actionCtx)

	return a
}

type actionEncryptionKeyAdd struct {
	actionImpl

	actionEmptyCheckProgress
}

func (a *actionEncryptionKeyAdd) Start(ctx context.Context) (bool, error) {
	if err := ensureEncryptionSupport(a.actionCtx); err != nil {
		a.log.Err(err).Error("Action not supported")
		return true, nil
	}

	secret := a.actionCtx.GetSpec().RocksDB.Encryption.GetKeySecretName()
	if s, ok := a.action.Params[secretActionParam]; ok {
		secret = s
	}

	sha, d, exists, err := pod.GetEncryptionKey(ctx, a.actionCtx.ACS().CurrentClusterCache().Secret().V1().Read(), secret)
	if err != nil {
		a.log.Err(err).Error("Unable to fetch current encryption key")
		return true, nil
	}

	if !exists {
		return true, nil
	}

	p := patch.NewPatch()
	p.ItemAdd(patch.NewPath("data", sha), base64.StdEncoding.EncodeToString(d))

	patch, err := p.Marshal()
	if err != nil {
		a.log.Err(err).Error("Unable to encrypt patch")
		return true, nil
	}

	err = globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
		_, err := a.actionCtx.ACS().CurrentClusterCache().SecretsModInterface().V1().Patch(ctxChild, pod.GetEncryptionFolderSecretName(a.actionCtx.GetAPIObject().GetName()), types.JSONPatchType, patch, meta.PatchOptions{})
		return err
	})
	if err != nil {
		return false, err
	}

	return true, nil
}
