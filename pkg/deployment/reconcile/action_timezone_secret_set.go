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

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	"github.com/arangodb/kube-arangodb/pkg/deployment/patch"
	"github.com/arangodb/kube-arangodb/pkg/deployment/pod"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
)

func newTimezoneSecretSetAction(action api.Action, actionCtx ActionContext) Action {
	a := &actionTimezoneSecretSet{}

	a.actionImpl = newActionImplDefRef(action, actionCtx)

	return a
}

type actionTimezoneSecretSet struct {
	actionImpl

	actionEmptyCheckProgress
}

func (a *actionTimezoneSecretSet) Start(ctx context.Context) (bool, error) {
	if !features.Timezone().Enabled() {
		return true, nil
	}

	secrets := a.actionCtx.ACS().CurrentClusterCache().Secret().V1()

	tz, ok := GetTimezone(a.actionCtx.GetSpec().Timezone)
	if !ok {
		return true, nil
	}

	tzd, ok := tz.GetData()
	if !ok {
		return true, nil
	}

	if IsTimezoneValid(secrets, a.actionCtx.GetName(), tz) {
		return true, nil
	}

	if s, ok := secrets.GetSimple(pod.TimezoneSecret(a.actionCtx.GetName())); ok {
		// Exists
		// We need to prepare patch
		data := map[string]string{}

		data[pod.TimezoneNameKey] = base64.StdEncoding.EncodeToString([]byte(tz.Name))
		data[pod.TimezoneDataKey] = base64.StdEncoding.EncodeToString(tzd)
		data[pod.TimezoneTZKey] = base64.StdEncoding.EncodeToString([]byte(tz.Name))

		p := patch.NewPatch()
		p.ItemReplace(patch.NewPath("data"), data)

		patch, err := p.Marshal()
		if err != nil {
			a.log.Err(err).Error("Unable to encrypt patch")
			return true, nil
		}

		err = globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
			_, err := a.actionCtx.ACS().CurrentClusterCache().SecretsModInterface().V1().Patch(ctxChild, s.GetName(), types.JSONPatchType, patch, meta.PatchOptions{})
			return err
		})
		if err != nil {
			if !kerrors.IsInvalid(err) {
				return false, errors.Wrapf(err, "Unable to update secret: %s", s.GetName())
			}
		}

		return true, nil
	} else {
		s = &core.Secret{
			ObjectMeta: meta.ObjectMeta{
				Name:      pod.TimezoneSecret(a.actionCtx.GetName()),
				Namespace: a.actionCtx.GetNamespace(),
				OwnerReferences: []meta.OwnerReference{
					a.actionCtx.GetAPIObject().AsOwner(),
				},
			},
			Data: map[string][]byte{
				pod.TimezoneNameKey: []byte(tz.Name),
				pod.TimezoneDataKey: tzd,
				pod.TimezoneTZKey:   []byte(tz.Zone),
			},
			Type: core.SecretTypeOpaque,
		}

		if err := globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
			_, err := a.actionCtx.ACS().CurrentClusterCache().SecretsModInterface().V1().Create(ctxChild, s, meta.CreateOptions{})
			return err
		}); err != nil {
			a.log.Err(err).Error("Unable to create cm secret")
			return true, nil
		}
	}

	return true, nil
}
