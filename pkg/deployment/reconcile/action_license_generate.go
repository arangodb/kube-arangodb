//
// DISCLAIMER
//
// Copyright 2025 ArangoDB GmbH, Cologne, Germany
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

	"google.golang.org/protobuf/types/known/durationpb"
	core "k8s.io/api/core/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/client"
	"github.com/arangodb/kube-arangodb/pkg/license_manager"
	"github.com/arangodb/kube-arangodb/pkg/platform/inventory"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	ugrpc "github.com/arangodb/kube-arangodb/pkg/util/grpc"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

func newLicenseGenerateAction(action api.Action, actionCtx ActionContext) Action {
	a := &actionLicenseGenerate{}

	a.actionImpl = newActionImplDefRef(action, actionCtx)

	return a
}

type actionLicenseGenerate struct {
	actionImpl

	actionEmptyCheckProgress
}

func (a *actionLicenseGenerate) Start(ctx context.Context) (bool, error) {
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

	if l.Master == nil {
		return true, nil
	}

	m, ok := a.actionCtx.GetMemberStatusByID(a.action.MemberID)
	if !ok {
		a.log.Error("No such member")
		return true, nil
	}

	c, err := a.actionCtx.GetMembersState().GetMemberClient(m.ID)
	if err != nil {
		a.log.Err(err).Error("Unable to get client")
		return true, nil
	}

	inv, err := inventory.FetchInventorySpec(ctx, a.log, 4, c.Connection(), &inventory.Configuration{Telemetry: util.NewType(true)})
	if err != nil {
		a.log.Err(err).Error("Unable to generate inventory")
		return true, nil
	}

	lm, err := license_manager.NewClient(license_manager.ArangoLicenseManagerEndpoint, l.Master.ClientID, l.Master.ClientSecret)
	if err != nil {
		a.log.Err(err).Error("Unable to create inventory client")
		return true, nil
	}

	license, err := lm.License(ctx, license_manager.LicenseRequest{
		DeploymentID: util.NewType(inv.DeploymentId),
		TTL:          util.NewType(ugrpc.NewObject(durationpb.New(spec.License.GetTTL()))),
		Inventory:    util.NewType(ugrpc.NewObject(inv)),
	})
	if err != nil {
		a.log.Err(err).Error("Unable to create license")
		a.actionCtx.CreateEvent(&k8sutil.Event{
			InvolvedObject: a.actionCtx.GetAPIObject(),
			Type:           core.EventTypeWarning,
			Reason:         "License Generation Failed",
			Message:        fmt.Sprintf("License Generation Failed with: %s", err),
		})
		return true, nil
	}
	a.log.Str("id", license.ID).Info("License Generated")

	if err := client.NewClient(c.Connection(), a.log).SetLicense(ctxChild, license.License, true); err != nil {
		a.log.Err(err).Error("Unable to set license")
		return true, nil
	}

	return true, nil
}
