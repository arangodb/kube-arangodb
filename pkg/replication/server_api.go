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

package replication

import (
	api "github.com/arangodb/kube-arangodb/pkg/apis/replication/v1"
	"github.com/arangodb/kube-arangodb/pkg/server"
)

// Name returns the name of the deployment.
func (dr *DeploymentReplication) Name() string {
	return dr.apiObject.Name
}

// Namespace returns the namespace that contains the deployment.
func (dr *DeploymentReplication) Namespace() string {
	return dr.apiObject.Namespace
}

// StateColor determinates the state of the deployment in color codes.
func (dr *DeploymentReplication) StateColor() server.StateColor {
	switch dr.status.Phase {
	case api.DeploymentReplicationPhaseFailed:
		return server.StateRed
	}
	if dr.status.Conditions.IsTrue(api.ConditionTypeConfigured) {
		return server.StateGreen
	}
	return server.StateYellow
}

// Source provides info on the source of the replication
func (dr *DeploymentReplication) Source() server.Endpoint {
	return serverEndpoint{
		dr:      dr,
		getSpec: func() api.EndpointSpec { return dr.apiObject.Spec.Source },
	}
}

// Destination provides info on the destination of the replication
func (dr *DeploymentReplication) Destination() server.Endpoint {
	return serverEndpoint{
		dr:      dr,
		getSpec: func() api.EndpointSpec { return dr.apiObject.Spec.Destination },
	}
}
