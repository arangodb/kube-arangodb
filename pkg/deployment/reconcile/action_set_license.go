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

	"github.com/arangodb/kube-arangodb/pkg/util/globals"

	"github.com/arangodb/kube-arangodb/pkg/deployment/client"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/rs/zerolog"
)

func init() {
	registerAction(api.ActionTypeLicenseSet, newLicenseSet)
}

func newLicenseSet(log zerolog.Logger, action api.Action, actionCtx ActionContext) Action {
	a := &licenseSetAction{}

	a.actionImpl = newActionImplDefRef(log, action, actionCtx, defaultTimeout)

	return a
}

type licenseSetAction struct {
	actionImpl

	actionEmptyCheckProgress
}

func (a *licenseSetAction) Start(ctx context.Context) (bool, error) {
	log := a.log

	spec := a.actionCtx.GetSpec()
	if !spec.License.HasSecretName() {
		log.Error().Msg("License is not set")
		return true, nil
	}

	l, ok := k8sutil.GetLicenseFromSecret(a.actionCtx.GetCachedStatus(), spec.License.GetSecretName())

	if !ok {
		return true, nil
	}

	if !l.V2.IsV2Set() {
		return true, nil
	}

	group := a.action.Group
	m, ok := a.actionCtx.GetMemberStatusByID(a.action.MemberID)
	if !ok {
		log.Error().Msg("No such member")
		return true, nil
	}

	ctxChild, cancel := globals.GetGlobals().Timeouts().ArangoD().WithTimeout(ctx)
	defer cancel()

	c, err := a.actionCtx.GetServerClient(ctxChild, group, m.ID)
	if !ok {
		log.Error().Err(err).Msg("Unable to get client")
		return true, nil
	}

	client := client.NewClient(c.Connection())

	if ok, err := licenseV2Compare(ctx, client, l.V2); err != nil {
		log.Error().Err(err).Msg("Unable to verify license")
		return true, nil
	} else if ok {
		// Already latest license
		return true, nil
	}

	if err := client.SetLicense(ctx, string(l.V2), true); err != nil {
		log.Error().Err(err).Msg("Unable to set license")
		return true, nil
	}

	return true, nil
}

func licenseV2Compare(ctx context.Context, client client.Client, license k8sutil.License) (bool, error) {
	currentLicense, err := client.GetLicense(ctx)
	if err != nil {
		return false, err
	}

	if currentLicense.Hash == license.V2Hash() {
		// Already latest license
		return true, nil
	}

	return false, nil
}
