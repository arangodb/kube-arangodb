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

	"github.com/arangodb/kube-arangodb/pkg/deployment/resources"
	"github.com/arangodb/kube-arangodb/pkg/util"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/rs/zerolog"
)

func init() {
	registerAction(api.ActionTypeTLSKeyStatusUpdate, newTLSKeyStatusUpdate)
}

func newTLSKeyStatusUpdate(log zerolog.Logger, action api.Action, actionCtx ActionContext) Action {
	a := &tlsKeyStatusUpdateAction{}

	a.actionImpl = newActionImplDefRef(log, action, actionCtx, defaultTimeout)

	return a
}

type tlsKeyStatusUpdateAction struct {
	actionImpl

	actionEmptyCheckProgress
}

func (a *tlsKeyStatusUpdateAction) Start(ctx context.Context) (bool, error) {
	if !a.actionCtx.GetSpec().TLS.IsSecure() {
		return true, nil
	}

	f, err := a.actionCtx.SecretsInterface().Get(resources.GetCASecretName(a.actionCtx.GetAPIObject()), meta.GetOptions{})
	if err != nil {
		a.log.Error().Err(err).Msgf("Unable to get folder info")
		return true, nil
	}

	keyHashes := secretKeysToListWithPrefix("sha256:", f)

	if err = a.actionCtx.WithStatusUpdate(func(s *api.DeploymentStatus) bool {
		if len(keyHashes) == 1 {
			if s.Hashes.TLS.CA == nil || *s.Hashes.TLS.CA != keyHashes[0] {
				s.Hashes.TLS.CA = util.NewString(keyHashes[0])
				return true
			}
		}

		if !util.CompareStringArray(keyHashes, s.Hashes.TLS.Truststore) {
			s.Hashes.TLS.Truststore = keyHashes
			return true
		}

		return false
	}); err != nil {
		return false, err
	}

	return true, nil
}
