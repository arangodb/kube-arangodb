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

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/client"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

func newLicenseSetAction(action api.Action, actionCtx ActionContext) Action {
	a := &actionLicenseSet{}

	a.actionImpl = newActionImplDefRef(action, actionCtx)

	return a
}

type actionLicenseSet struct {
	actionImpl

	actionEmptyCheckProgress
}

func (a *actionLicenseSet) Start(ctx context.Context) (bool, error) {
	ctxChild, cancel := globals.GetGlobals().Timeouts().ArangoD().WithTimeout(ctx)
	defer cancel()
	spec := a.actionCtx.GetSpec()
	if !spec.License.HasSecretName() {
		a.log.Error("License is not set")
		return true, nil
	}

	l, err := k8sutil.GetLicenseFromSecret(a.actionCtx.ACS().CurrentClusterCache(), spec.License.GetSecretName())
	if err != nil {
		return true, err
	}

	if !l.V2.IsV2Set() {
		return true, nil
	}

	m, ok := a.actionCtx.GetMemberStatusByID(a.action.MemberID)
	if !ok {
		a.log.Error("No such member")
		return true, nil
	}

	c, err := a.actionCtx.GetMembersState().GetMemberClient(m.ID)
	if !ok {
		a.log.Err(err).Error("Unable to get client")
		return true, nil
	}

	client := client.NewClient(c.Connection(), a.log)

	if ok, err := licenseV2Compare(ctxChild, client, l.V2); err != nil {
		a.log.Err(err).Error("Unable to verify license")
		return true, nil
	} else if ok {
		// Already latest license
		return true, nil
	}

	if err := client.SetLicense(ctxChild, string(l.V2), true); err != nil {
		a.log.Err(err).Error("Unable to set license")
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
