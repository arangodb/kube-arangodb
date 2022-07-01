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

package server

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/arangodb/kube-arangodb/pkg/logging"
)

var serverLogger = logging.Global().RegisterAndGetLogger("server", logging.Info)

type operatorsResponse struct {
	PodName               string              `json:"pod"`
	Namespace             string              `json:"namespace"`
	Deployment            bool                `json:"deployment"`
	DeploymentReplication bool                `json:"deployment_replication"`
	Storage               bool                `json:"storage"`
	Other                 []OperatorReference `json:"other"`
}

type OperatorType string

const (
	OperatorTypeDeployment            OperatorType = "deployment"
	OperatorTypeDeploymentReplication OperatorType = "deployment_replication"
	OperatorTypeStorage               OperatorType = "storage"
)

// OperatorReference contains a reference to another operator
type OperatorReference struct {
	Namespace string       `json:"namespace"`
	Type      OperatorType `json:"type"`
	URL       string       `json:"url"`
}

// Handle a GET /api/operators request
func (s *Server) handleGetOperators(c *gin.Context) {
	result := operatorsResponse{
		PodName:               s.cfg.PodName,
		Namespace:             s.cfg.Namespace,
		Deployment:            s.deps.Deployment.Probe.IsReady(),
		DeploymentReplication: s.deps.DeploymentReplication.Probe.IsReady(),
		Storage:               s.deps.Storage.Probe.IsReady(),
		Other:                 s.deps.Operators.FindOtherOperators(),
	}
	serverLogger.Interface("result", result).Info("handleGetOperators")
	c.JSON(http.StatusOK, result)
}
