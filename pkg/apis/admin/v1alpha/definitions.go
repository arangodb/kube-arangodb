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
package v1alpha

import (
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ConditionType string

const (
	// ConditionTypeReady the resource has been created is and is ready for use
	ConditionTypeReady ConditionType = "ready"
	// ConditionTypeDeleted the resource has been deleted
	ConditionTypeDeleted ConditionType = "deleted"
	// ConditionTypeFailed creating the resource has failed
	ConditionTypeFailed ConditionType = "failed"
)

type Condition struct {
	LastUpdated metav1.Time
	Status      v1.ConditionStatus
	Reason      string
	Message     string
}

type ConditionList map[ConditionType]Condition

// SetCondition sets the condition on this condition list
func (cl *ConditionList) SetCondition(condition ConditionType, status v1.ConditionStatus, reason, message string) {
	if *cl == nil {
		*cl = make(ConditionList)
	}

	(*cl)[condition] = Condition{
		LastUpdated: metav1.Now(),
		Status:      status,
		Reason:      reason,
		Message:     message,
	}
}

// RemoveCondition removes the condition from the condition list
func (cl ConditionList) RemoveCondition(condition ConditionType) {
	delete(cl, condition)
}

type ResourceStatus struct {
	CreatedAt  *metav1.Time  `json:"createdAt,omitempty"`
	Conditions ConditionList `json:"conditions,omitempty"`
}
