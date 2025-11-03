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
	"math"
	"time"

	"google.golang.org/protobuf/types/known/durationpb"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

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

	if l.API == nil {
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

	var req license_manager.LicenseRequest
	did, err := inventory.ExtractDeploymentID(ctx, c.Connection())
	if err != nil {
		a.log.Err(err).Error("Unable to get deployment id")
		return true, nil
	}

	req.DeploymentID = util.NewType(did)

	if spec.License.GetInventory() {
		inv, err := inventory.FetchInventorySpec(ctx, a.log, 4, c.Connection(), &inventory.Configuration{Telemetry: util.NewType(spec.License.GetTelemetry())})
		if err != nil {
			a.log.Err(err).Error("Unable to generate inventory")
			return true, nil
		}

		if inv.DeploymentId != did {
			a.log.Err(err).Error("Invalid deployment ID in inventory")
			return true, nil
		}

		req.Inventory = util.NewType(ugrpc.NewObject(inv))
	}

	if q := spec.License.TTL; q != nil {
		req.TTL = util.NewType(ugrpc.NewObject(durationpb.New(q.Duration)))
	}

	lm, err := license_manager.NewClient(license_manager.ArangoLicenseManagerEndpoint, l.API.ClientID, l.API.ClientSecret)
	if err != nil {
		a.log.Err(err).Error("Unable to create inventory client")
		return true, nil
	}

	generatedLicense, err := lm.License(ctx, req)
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
	a.log.Str("id", generatedLicense.ID).Info("License Generated")

	client := client.NewClient(c.Connection(), a.log)

	if err := client.SetLicense(ctxChild, generatedLicense.License, true); err != nil {
		a.log.Err(err).Error("Unable to set license")
		return true, nil
	}

	license, err := client.GetLicense(ctxChild)
	if err != nil {
		a.log.Err(err).Error("Unable to get license")
		return true, nil
	}

	expiration := time.Until(license.Expires())
	if expiration <= 0 {
		a.log.Error("Unable to get license - invalid timestamp")
		return true, nil
	}

	if q := spec.License.ExpirationGracePeriod; q != nil {
		expiration = expiration - q.Duration
	} else {
		expiration = time.Duration(math.Round(api.LicenseExpirationGraceRatio * float64(expiration)))
	}
	if expiration <= 0 {
		a.log.Error("Unable to get license - invalid after evaluation")
		return true, nil
	}

	expires := time.Now().Add(expiration)

	if expires.After(license.Expires()) {
		// License will expire before grace period, reduce to 90%
		expires = time.Now().Add(time.Duration(math.Round(float64(time.Since(license.Expires())) * api.LicenseExpirationGraceRatio)))
	}

	if err := a.actionCtx.WithStatusUpdate(ctx, func(s *api.DeploymentStatus) bool {
		s.License = &api.DeploymentStatusLicense{
			ID:         generatedLicense.ID,
			Hash:       license.Hash,
			Expires:    meta.Time{Time: license.Expires()},
			Mode:       api.LicenseModeAPI,
			Regenerate: meta.Time{Time: expires},
		}
		return true
	}); err != nil {
		a.log.Err(err).Error("Unable to register license")
		return true, nil
	}
	return true, nil
}
