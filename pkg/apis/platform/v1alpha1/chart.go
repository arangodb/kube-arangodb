//
// DISCLAIMER
//
// Copyright 2024-2025 ArangoDB GmbH, Cologne, Germany
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

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ArangoPlatformChartList is a list of ArangoPlatform Chart.
type ArangoPlatformChartList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`

	Items []ArangoPlatformChart `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ArangoPlatformChart contains definition and status of the ArangoPlatform Chart.
type ArangoPlatformChart struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`

	Spec   ArangoPlatformChartSpec   `json:"spec"`
	Status ArangoPlatformChartStatus `json:"status"`
}

// AsOwner creates an OwnerReference for the given Extension
func (a *ArangoPlatformChart) AsOwner() meta.OwnerReference {
	trueVar := true
	return meta.OwnerReference{
		APIVersion: SchemeGroupVersion.String(),
		Kind:       platform.ArangoPlatformChartResourceKind,
		Name:       a.Name,
		UID:        a.UID,
		Controller: &trueVar,
	}
}

func (a *ArangoPlatformChart) GetStatus() ArangoPlatformChartStatus {
	return a.Status
}

func (a *ArangoPlatformChart) SetStatus(status ArangoPlatformChartStatus) {
	a.Status = status
}

func (a *ArangoPlatformChart) Ready() bool {
	if a == nil {
		return false
	}

	if a.Status.Info == nil {
		return false
	}

	if a.Status.Info.Details == nil {
		return false
	}

	if !a.Status.Conditions.IsTrue(ReadyCondition) {
		return false
	}

	if a.Status.Info.Checksum != a.Spec.Checksum() {
		return false
	}

	return true
}
