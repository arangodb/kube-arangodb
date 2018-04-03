//
// DISCLAIMER
//
// Copyright 2018 ArangoDB GmbH, Cologne, Germany
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
// Author Ewout Prangsma
//

package arangod

import (
	"context"

	driver "github.com/arangodb/go-driver"
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
func GetNumberOfServers(ctx context.Context, conn driver.Connection) (NumberOfServers, error) {
	req, err := conn.NewRequest("GET", "_admin/cluster/numberOfServers")
	if err != nil {
		return NumberOfServers{}, maskAny(err)
	}
	resp, err := conn.Do(ctx, req)
	if err != nil {
		return NumberOfServers{}, maskAny(err)
	}
	if err := resp.CheckStatus(200); err != nil {
		return NumberOfServers{}, maskAny(err)
	}
	var result NumberOfServers
	if err := resp.ParseBody("", &result); err != nil {
		return NumberOfServers{}, maskAny(err)
	}
	return result, nil
}

// SetNumberOfServers updates the number of servers the cluster has.
func SetNumberOfServers(ctx context.Context, conn driver.Connection, noCoordinators, noDBServers int) error {
	req, err := conn.NewRequest("PUT", "_admin/cluster/numberOfServers")
	if err != nil {
		return maskAny(err)
	}
	input := NumberOfServers{
		Coordinators: &noCoordinators,
		DBServers:    &noDBServers,
	}
	if _, err := req.SetBody(input); err != nil {
		return maskAny(err)
	}
	resp, err := conn.Do(ctx, req)
	if err != nil {
		return maskAny(err)
	}
	if err := resp.CheckStatus(200); err != nil {
		return maskAny(err)
	}
	return nil
}

// RemoveServerFromCluster tries to remove a coordinator or DBServer from the cluster.
func RemoveServerFromCluster(ctx context.Context, conn driver.Connection, id driver.ServerID) error {
	req, err := conn.NewRequest("POST", "_admin/cluster/removeServer")
	if err != nil {
		return maskAny(err)
	}
	if _, err := req.SetBody(id); err != nil {
		return maskAny(err)
	}
	resp, err := conn.Do(ctx, req)
	if err != nil {
		return maskAny(err)
	}
	if err := resp.CheckStatus(200); err != nil {
		return maskAny(err)
	}
	return nil
}
