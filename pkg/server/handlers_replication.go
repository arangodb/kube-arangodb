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

package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// DeploymentReplication is the API implemented by an ArangoDeploymentReplication.
type DeploymentReplication interface {
	Name() string
	Namespace() string
	StateColor() StateColor
}

// DeploymentReplicationOperator is the API implemented by the deployment operator.
type DeploymentReplicationOperator interface {
	// GetDeploymentReplications returns basic information for all deployment replications managed by the operator
	GetDeploymentReplications() ([]DeploymentReplication, error)
	// GetDeploymentReplication returns detailed information for a deployment replication, managed by the operator, with given name
	GetDeploymentReplication(name string) (DeploymentReplication, error)
}

// DeploymentReplicationInfo is the information returned per deployment replication.
type DeploymentReplicationInfo struct {
	Name       string     `json:"name"`
	Namespace  string     `json:"namespace"`
	StateColor StateColor `json:"state_color"`
}

// newDeploymentReplicationInfo initializes a DeploymentReplicationInfo for the given deployment replication.
func newDeploymentReplicationInfo(dr DeploymentReplication) DeploymentReplicationInfo {
	return DeploymentReplicationInfo{
		Name:       dr.Name(),
		Namespace:  dr.Namespace(),
		StateColor: dr.StateColor(),
	}
}

// DeploymentReplicationInfoDetails is the detailed information returned per deployment replication.
type DeploymentReplicationInfoDetails struct {
	DeploymentReplicationInfo
}

// newDeploymentReplicationInfoDetails initializes a DeploymentReplicationInfoDetails for the given deployment replication.
func newDeploymentReplicationInfoDetails(dr DeploymentReplication) DeploymentReplicationInfoDetails {
	result := DeploymentReplicationInfoDetails{
		DeploymentReplicationInfo: newDeploymentReplicationInfo(dr),
	}
	return result
}

// Handle a GET /api/deployment-replication request
func (s *Server) handleGetDeploymentReplications(c *gin.Context) {
	if do := s.deps.Operators.DeploymentReplicationOperator(); do != nil {
		// Fetch deployment replications
		repls, err := do.GetDeploymentReplications()
		if err != nil {
			sendError(c, err)
		} else {
			result := make([]DeploymentReplicationInfo, len(repls))
			for i, dr := range repls {
				result[i] = newDeploymentReplicationInfo(dr)
			}
			c.JSON(http.StatusOK, gin.H{
				"replications": result,
			})
		}
	}
}

// Handle a GET /api/deployment-replication/:name request
func (s *Server) handleGetDeploymentReplicationDetails(c *gin.Context) {
	if do := s.deps.Operators.DeploymentReplicationOperator(); do != nil {
		// Fetch deployments
		dr, err := do.GetDeploymentReplication(c.Params.ByName("name"))
		if err != nil {
			sendError(c, err)
		} else {
			result := newDeploymentReplicationInfoDetails(dr)
			c.JSON(http.StatusOK, result)
		}
	}
}
