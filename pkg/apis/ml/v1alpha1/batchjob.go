//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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
	"strings"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/apis/ml"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ArangoMLBatchJobList is a list of ArangoML BatchJobs.
type ArangoMLBatchJobList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`

	Items []ArangoMLBatchJob `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ArangoMLBatchJob contains the definition and status of the ArangoML BatchJob.
type ArangoMLBatchJob struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`

	Spec   ArangoMLBatchJobSpec   `json:"spec"`
	Status ArangoMLBatchJobStatus `json:"status"`
}

// AsOwner creates an OwnerReference for the given BatchJob
func (a *ArangoMLBatchJob) AsOwner() meta.OwnerReference {
	trueVar := true
	return meta.OwnerReference{
		APIVersion: SchemeGroupVersion.String(),
		Kind:       ml.ArangoMLBatchJobResourceKind,
		Name:       a.Name,
		UID:        a.UID,
		Controller: &trueVar,
	}
}

func (a *ArangoMLBatchJob) GetStatus() ArangoMLBatchJobStatus {
	return a.Status
}

func (a *ArangoMLBatchJob) SetStatus(status ArangoMLBatchJobStatus) {
	a.Status = status
}

func (a *ArangoMLBatchJob) GetJobType() string {
	val, ok := a.Labels[constants.MLJobTypeLabel]
	if !ok {
		return ""
	}
	return strings.ToLower(val)
}

func (a *ArangoMLBatchJob) GetScheduleType() string {
	val, ok := a.Labels[constants.MLJobScheduleLabel]
	if !ok {
		return ""
	}
	return strings.ToLower(val)
}

func (a *ArangoMLBatchJob) GetMLDeploymentName() string {
	val, ok := a.Labels[constants.MLJobScheduleLabel]
	if !ok {
		return ""
	}
	return val
}

func (a *ArangoMLBatchJob) ValidateLabels() error {
	depl, ok := a.Labels[constants.MLDeploymentLabel]
	if !ok {
		return errors.Newf("Job missing label: %s", constants.MLDeploymentLabel)
	}
	if depl == "" {
		return errors.Newf("Job empty value for label: %s", constants.MLDeploymentLabel)
	}

	t, ok := a.Labels[constants.MLJobTypeLabel]
	if !ok {
		return errors.Newf("Job missing label: %s", constants.MLJobTypeLabel)
	}
	jobType := strings.ToLower(t)
	if jobType != constants.MLJobTrainingType && jobType != constants.MLJobPredictionType {
		return errors.Newf("Job label (%s) has unexpected value: %s", constants.MLJobTypeLabel, t)
	}

	s, ok := a.Labels[constants.MLJobScheduleLabel]
	if !ok {
		return errors.Newf("Job missing label: %s", constants.MLJobTypeLabel)
	}
	scheduleType := strings.ToLower(s)
	if scheduleType != constants.MLJobScheduleCPU && scheduleType != constants.MLJobScheduleGPU {
		return errors.Newf("Job label (%s) has unexpected value: %s", constants.MLJobScheduleLabel, s)
	}

	return nil
}
