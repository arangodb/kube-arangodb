//
// DISCLAIMER
//
// Copyright 2016-2026 ArangoDB GmbH, Cologne, Germany
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
	"github.com/arangodb/kube-arangodb/pkg/deployment/versions"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
)

type actionEnableMemberMaintenance struct {
	// actionImpl implement timeout and member id functions
	actionImpl
}

func (a actionEnableMemberMaintenance) Start(ctx context.Context) (bool, error) {
	if a.action.Group != api.ServerGroupDBServers {
		return true, nil
	}

	m, ok := a.actionCtx.GetMemberStatusByID(a.action.MemberID)
	if !ok {
		a.log.Error("No such member")
		return true, nil
	}

	info, ok := memberImageInfo(a.actionCtx.GetSpec(), m, a.actionCtx.GetStatus().Images)
	if !ok {
		a.log.Error("Unable to get image info")
		return true, nil
	}

	if !versions.MemberMaintenance(info) {
		a.log.Error("MemberMaintenance feature not ready for version")
		return true, nil
	}

	cache, ok := a.actionCtx.GetAgencyCache()
	if !ok {
		a.log.Debug("AgencyCache is not ready")
		return true, nil
	} else if cache.Supervision.Maintenance.Exists() {
		a.log.Debug("Cluster Maintenance mode is enabled")
		return true, nil
	}

	databaseClient, err := a.actionCtx.GetMembersState().State().GetDatabaseClient()
	if err != nil {
		a.log.Err(err).Error("Unable to get client")
		return true, nil
	}

	nctx, c := globals.GetGlobalTimeouts().ArangoD().WithTimeout(ctx)
	defer c()

	if err := client.NewClient(databaseClient.Connection()).EnableMaintenanceWithDefaultTimeout(nctx, m.ID); err != nil {
		a.log.Err(err).Error("Unable to enable maintenance")
		return true, nil
	}

	return true, nil
}

func (a actionEnableMemberMaintenance) CheckProgress(_ context.Context) (bool, bool, error) {
	m, ok := a.actionCtx.GetMemberStatusByID(a.action.MemberID)
	if !ok {
		a.log.Error("No such member")
		return true, false, nil
	}

	if m.Conditions.IsTrue(api.ConditionTypeMemberMaintenanceMode) {
		return true, false, nil
	}

	return false, false, nil
}
