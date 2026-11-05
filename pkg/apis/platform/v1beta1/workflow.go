//
// DISCLAIMER
//
// Copyright 2026 ArangoDB GmbH, Cologne, Germany
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
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/apis/platform"
)

const (
	FinalizerArangoPlatformWorkflowRelease = platform.ArangoPlatformWorkflowCRDName + "/cleanup"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ArangoPlatformWorkflowList is a list of ArangoPlatform Workflow.
type ArangoPlatformWorkflowList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`

	Items []ArangoPlatformWorkflow `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ArangoPlatformWorkflow contains definition and status of the ArangoPlatform Workflow.
type ArangoPlatformWorkflow struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`

	Spec   ArangoPlatformWorkflowSpec   `json:"spec"`
	Status ArangoPlatformWorkflowStatus `json:"status"`
}

// AsOwner creates an OwnerReference for the given Extension
func (a *ArangoPlatformWorkflow) AsOwner() meta.OwnerReference {
	trueVar := true
	return meta.OwnerReference{
		APIVersion: SchemeGroupVersion.String(),
		Kind:       platform.ArangoPlatformWorkflowResourceKind,
		Name:       a.Name,
		UID:        a.UID,
		Controller: &trueVar,
	}
}

func (a *ArangoPlatformWorkflow) GetStatus() ArangoPlatformWorkflowStatus {
	return a.Status
}

func (a *ArangoPlatformWorkflow) SetStatus(status ArangoPlatformWorkflowStatus) {
	a.Status = status
}
