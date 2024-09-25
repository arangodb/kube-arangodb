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
	batch "k8s.io/api/batch/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/apis/scheduler"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ArangoSchedulerCronJobList is a list of CronJobs.
type ArangoSchedulerCronJobList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`

	Items []ArangoSchedulerCronJob `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ArangoSchedulerCronJob wraps batch. ArangoSchedulerCronJob with profile details
type ArangoSchedulerCronJob struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`

	Spec   ArangoSchedulerCronJobSpec   `json:"spec"`
	Status ArangoSchedulerCronJobStatus `json:"status"`
}

type ArangoSchedulerCronJobSpec struct {
	// Profiles keeps list of the profiles
	Profiles []string `json:"profiles,omitempty"`

	batch.CronJobSpec `json:",inline"`
}

type ArangoSchedulerCronJobStatus struct {
	ArangoSchedulerStatusMetadata `json:",inline"`

	batch.CronJobStatus `json:",inline"`
}

// AsOwner creates an OwnerReference for the given  ArangoSchedulerCronJob
func (d *ArangoSchedulerCronJob) AsOwner() meta.OwnerReference {
	trueVar := true
	return meta.OwnerReference{
		APIVersion: SchemeGroupVersion.String(),
		Kind:       scheduler.CronJobResourceKind,
		Name:       d.Name,
		UID:        d.UID,
		Controller: &trueVar,
	}
}

func (d *ArangoSchedulerCronJob) GetStatus() ArangoSchedulerCronJobStatus {
	return d.Status
}

func (d *ArangoSchedulerCronJob) SetStatus(status ArangoSchedulerCronJobStatus) {
	d.Status = status
}
