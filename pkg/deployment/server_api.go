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

package deployment

import (
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
)

// Name returns the name of the deployment.
func (d *Deployment) Name() string {
	return d.apiObject.Name
}

// Mode returns the mode of the deployment.
func (d *Deployment) Mode() api.DeploymentMode {
	return d.GetSpec().GetMode()
}

// PodCount returns the number of pods for the deployment
func (d *Deployment) PodCount() int {
	count := 0
	status, _ := d.GetStatus()
	status.Members.ForeachServerGroup(func(group api.ServerGroup, list api.MemberStatusList) error {
		for _, m := range list {
			if m.PodName != "" {
				count++
			}
		}
		return nil
	})
	return count
}

// ReadyPodCount returns the number of pods for the deployment that are in ready state
func (d *Deployment) ReadyPodCount() int {
	count := 0
	status, _ := d.GetStatus()
	status.Members.ForeachServerGroup(func(group api.ServerGroup, list api.MemberStatusList) error {
		for _, m := range list {
			if m.PodName == "" {
				continue
			}
			if m.Conditions.IsTrue(api.ConditionTypeReady) {
				count++
			}
		}
		return nil
	})
	return count
}
