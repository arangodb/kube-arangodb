//
// DISCLAIMER
//
// Copyright 2025 ArangoDB GmbH, Cologne, Germany
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

package v1alpha1

import (
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/apis/platform"
)

const (
	FinalizerArangoPlatformServiceRelease = platform.ArangoPlatformServiceCRDName + "/cleanup"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ArangoPlatformServiceList is a list of ArangoPlatform Service.
type ArangoPlatformServiceList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`

	Items []ArangoPlatformService `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ArangoPlatformService contains definition and status of the ArangoPlatform Service.
type ArangoPlatformService struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`

	Spec   ArangoPlatformServiceSpec   `json:"spec"`
	Status ArangoPlatformServiceStatus `json:"status"`
}

// AsOwner creates an OwnerReference for the given Extension
func (a *ArangoPlatformService) AsOwner() meta.OwnerReference {
	trueVar := true
	return meta.OwnerReference{
		APIVersion: SchemeGroupVersion.String(),
		Kind:       platform.ArangoPlatformServiceResourceKind,
		Name:       a.Name,
		UID:        a.UID,
		Controller: &trueVar,
	}
}

func (a *ArangoPlatformService) GetStatus() ArangoPlatformServiceStatus {
	return a.Status
}

func (a *ArangoPlatformService) SetStatus(status ArangoPlatformServiceStatus) {
	a.Status = status
}
