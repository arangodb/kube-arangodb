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

package arangod

import (
	"context"
	goHttp "net/http"

	adbDriverV2 "github.com/arangodb/go-driver/v2/arangodb"
	adbDriverV2Connection "github.com/arangodb/go-driver/v2/connection"
)

// NumberOfServers is the JSON structure return for the numberOfServers API call.
type NumberOfServers struct {
	Coordinators *int `json:"numberOfCoordinators,omitempty"`
	DBServers    *int `json:"numberOfDBServers,omitempty"`
}

// GetCoordinators returns Coordinators if not nil, otherwise 0.
func (n NumberOfServers) GetCoordinators() int {
	if n.Coordinators != nil {
		return *n.Coordinators
	}
	return 0
}

// GetDBServers returns DBServers if not nil, otherwise 0.
func (n NumberOfServers) GetDBServers() int {
	if n.DBServers != nil {
		return *n.DBServers
	}
	return 0
}

// GetNumberOfServers fetches the number of servers the cluster wants to have.
func GetNumberOfServers(ctx context.Context, conn adbDriverV2Connection.Connection) (NumberOfServers, error) {
	return GetRequest[NumberOfServers](ctx, conn, "_admin/cluster/numberOfServers").AcceptCode(goHttp.StatusOK).Response()
}

// SetNumberOfServers updates the number of servers the cluster has.
func SetNumberOfServers(ctx context.Context, conn adbDriverV2Connection.Connection, noCoordinators, noDBServers *int) error {
	return PutRequest[NumberOfServers, any](ctx, conn, NumberOfServers{
		Coordinators: noCoordinators,
		DBServers:    noDBServers,
	}, "_admin/cluster/numberOfServers").AcceptCode(goHttp.StatusOK).Evaluate()
}

// CleanNumberOfServers removes the server count
func CleanNumberOfServers(ctx context.Context, conn adbDriverV2Connection.Connection) error {
	return PutRequest[map[string]interface{}, any](ctx, conn, map[string]interface{}{
		"numberOfCoordinators": nil,
		"numberOfDBServers":    nil,
	}, "_admin/cluster/numberOfServers").AcceptCode(goHttp.StatusOK).Evaluate()
}

// RemoveServerFromCluster tries to remove a coordinator or DBServer from the cluster.
func RemoveServerFromCluster(ctx context.Context, conn adbDriverV2Connection.Connection, id adbDriverV2.ServerID) error {
	return PostRequest[adbDriverV2.ServerID, any](ctx, conn, id, "_admin/cluster/removeServer").AcceptCode(goHttp.StatusOK).Evaluate()
}
