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

package v2alpha1

import "k8s.io/apimachinery/pkg/types"

type ArangoClusterSynchronizationStatus struct {
	Deployment       *ArangoClusterSynchronizationDeploymentStatus `json:"deployment,omitempty"`
	RemoteDeployment *ArangoClusterSynchronizationDeploymentStatus `json:"remoteDeployment,omitempty"`

	Conditions ConditionList `json:"conditions,omitempty"`
}

type ArangoClusterSynchronizationDeploymentStatus struct {
	Name      string    `json:"name"`
	Namespace string    `json:"namespace"`
	UID       types.UID `json:"uid"`
}
