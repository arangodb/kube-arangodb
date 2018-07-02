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

package operator

import (
	"sort"

	"github.com/arangodb/kube-arangodb/pkg/server"
)

// DeploymentOperator provides access to the deployment operator.
func (o *Operator) DeploymentOperator() server.DeploymentOperator {
	return o
}

// GetDeployments returns all current deployments
func (o *Operator) GetDeployments() ([]server.Deployment, error) {
	o.Dependencies.LivenessProbe.Lock()
	defer o.Dependencies.LivenessProbe.Unlock()

	result := make([]server.Deployment, 0, len(o.deployments))
	for _, d := range o.deployments {
		result = append(result, d)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name() < result[j].Name()
	})
	return result, nil
}
