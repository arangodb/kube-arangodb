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

package operator

import (
	"sort"

	"github.com/arangodb/kube-arangodb/pkg/server"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
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

// GetDeployment returns detailed information for a deployment, managed by the operator, with given name
func (o *Operator) GetDeployment(name string) (server.Deployment, error) {
	o.Dependencies.LivenessProbe.Lock()
	defer o.Dependencies.LivenessProbe.Unlock()

	for _, d := range o.deployments {
		if d.Name() == name {
			return d, nil
		}
	}
	return nil, errors.WithStack(server.NotFoundError)
}

// DeploymentReplicationOperator provides access to the deployment replication operator.
func (o *Operator) DeploymentReplicationOperator() server.DeploymentReplicationOperator {
	return o
}

// GetDeploymentReplications returns all current deployments
func (o *Operator) GetDeploymentReplications() ([]server.DeploymentReplication, error) {
	o.Dependencies.LivenessProbe.Lock()
	defer o.Dependencies.LivenessProbe.Unlock()

	result := make([]server.DeploymentReplication, 0, len(o.deploymentReplications))
	for _, d := range o.deploymentReplications {
		result = append(result, d)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name() < result[j].Name()
	})
	return result, nil
}

// GetDeploymentReplication returns detailed information for a deployment replication, managed by the operator, with given name
func (o *Operator) GetDeploymentReplication(name string) (server.DeploymentReplication, error) {
	o.Dependencies.LivenessProbe.Lock()
	defer o.Dependencies.LivenessProbe.Unlock()

	for _, d := range o.deploymentReplications {
		if d.Name() == name {
			return d, nil
		}
	}
	return nil, errors.WithStack(server.NotFoundError)
}

// StorageOperator provides the local storage operator (if any)
func (o *Operator) StorageOperator() server.StorageOperator {
	return o
}

// GetLocalStorages returns basic information for all local storages managed by the operator
func (o *Operator) GetLocalStorages() ([]server.LocalStorage, error) {
	o.Dependencies.LivenessProbe.Lock()
	defer o.Dependencies.LivenessProbe.Unlock()

	result := make([]server.LocalStorage, 0, len(o.localStorages))
	for _, ls := range o.localStorages {
		result = append(result, ls)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name() < result[j].Name()
	})
	return result, nil
}

// GetLocalStorage returns detailed information for a local, managed by the operator, with given name
func (o *Operator) GetLocalStorage(name string) (server.LocalStorage, error) {
	o.Dependencies.LivenessProbe.Lock()
	defer o.Dependencies.LivenessProbe.Unlock()

	for _, ls := range o.localStorages {
		if ls.Name() == name {
			return ls, nil
		}
	}
	return nil, errors.WithStack(server.NotFoundError)
}
