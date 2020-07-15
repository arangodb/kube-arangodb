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
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"

	"github.com/arangodb/kube-arangodb/pkg/deployment/patch"
	"github.com/arangodb/kube-arangodb/pkg/deployment/pod"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/rs/zerolog"
)

func ensureEncryptionSupport(actionCtx ActionContext) error {
	if !actionCtx.GetSpec().RocksDB.IsEncrypted() {
		return errors.Errorf("Encryption is disabled")
	}

	if image, ok := actionCtx.GetCurrentImageInfo(); !ok {
		return errors.Errorf("Missing image info")
	} else {
		if !features.EncryptionRotation().Supported(image.ArangoDBVersion, image.Enterprise) {
			return errors.Errorf("Supported only in Enterprise Edition 3.7.0+")
		}
	}
	return nil
}

func init() {
	registerAction(api.ActionTypeEncryptionKeyAdd, newEncryptionKeyAdd)
}

func newEncryptionKeyAdd(log zerolog.Logger, action api.Action, actionCtx ActionContext) Action {
	a := &encryptionKeyAddAction{}

	a.actionImpl = newActionImplDefRef(log, action, actionCtx, defaultTimeout)

	return a
}

type encryptionKeyAddAction struct {
	actionImpl

	actionEmptyCheckProgress
}

func (a *encryptionKeyAddAction) Start(ctx context.Context) (bool, error) {
	if err := ensureEncryptionSupport(a.actionCtx); err != nil {
		a.log.Error().Err(err).Msgf("Action not supported")
		return true, nil
	}

	secret := a.actionCtx.GetSpec().RocksDB.Encryption.GetKeySecretName()
	if s, ok := a.action.Params[secretActionParam]; ok {
		secret = s
	}

	sha, d, exists, err := pod.GetEncryptionKey(a.actionCtx.SecretsInterface(), secret)
	if err != nil {
		a.log.Error().Err(err).Msgf("Unable to fetch current encryption key")
		return true, nil
	}

	if !exists {
		return true, nil
	}

	p := patch.NewPatch()
	p.ItemAdd(patch.NewPath("data", sha), base64.StdEncoding.EncodeToString(d))

	patch, err := p.Marshal()
	if err != nil {
		a.log.Error().Err(err).Msgf("Unable to encrypt patch")
		return true, nil
	}

	_, err = a.actionCtx.SecretsInterface().Patch(pod.GetEncryptionFolderSecretName(a.actionCtx.GetAPIObject().GetName()), types.JSONPatchType, patch)
	if err != nil {
		return false, err
	}

	return true, nil
}
