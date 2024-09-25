//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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

package v1beta1

import (
	apps "k8s.io/api/apps/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/apis/scheduler"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ArangoSchedulerDeploymentList is a list of Deployments.
type ArangoSchedulerDeploymentList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`

	Items []ArangoSchedulerDeployment `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ArangoSchedulerDeployment wraps apps. ArangoSchedulerDeployment with profile details
type ArangoSchedulerDeployment struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`

	Spec   ArangoSchedulerDeploymentSpec   `json:"spec"`
	Status ArangoSchedulerDeploymentStatus `json:"status"`
}

type ArangoSchedulerDeploymentSpec struct {
	// Profiles keeps list of the profiles
	Profiles []string `json:"profiles,omitempty"`

	apps.DeploymentSpec `json:",inline"`
}

type ArangoSchedulerDeploymentStatus struct {
	ArangoSchedulerStatusMetadata `json:",inline"`

	apps.DeploymentStatus `json:",inline"`
}

// AsOwner creates an OwnerReference for the given  ArangoSchedulerDeployment
func (d *ArangoSchedulerDeployment) AsOwner() meta.OwnerReference {
	trueVar := true
	return meta.OwnerReference{
		APIVersion: SchemeGroupVersion.String(),
		Kind:       scheduler.DeploymentResourceKind,
		Name:       d.Name,
		UID:        d.UID,
		Controller: &trueVar,
	}
}

func (d *ArangoSchedulerDeployment) GetStatus() ArangoSchedulerDeploymentStatus {
	return d.Status
}

func (d *ArangoSchedulerDeployment) SetStatus(status ArangoSchedulerDeploymentStatus) {
	d.Status = status
}
