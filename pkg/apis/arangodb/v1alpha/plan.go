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

package v1alpha

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// ActionType is a strongly typed name for a plan action item
type ActionType string

const (
	// ActionTypeAddMember causes a member to be added.
	ActionTypeAddMember ActionType = "AddMember"
	// ActionTypeRemoveMember causes a member to be removed.
	ActionTypeRemoveMember ActionType = "RemoveMember"
	// ActionTypeDrainMember causes a member to be drained (dbserver only).
	ActionTypeDrainMember ActionType = "DrainMember"
	// ActionTypeShutdownMember causes a member to be shutdown.
	ActionTypeShutdownMember ActionType = "ShutdownMember"
)

// Action represents a single action to be taken to update a deployment.
type Action struct {
	// Type of action.
	Type ActionType `json:"type"`
	// ID reference of the member involved in this action (if any)
	MemberID string `json:"memberID,omitempty"`
	// Group involved in this action
	Group ServerGroup `json:"group,omitempty"`
	// StartTime is set the when the action has been started, but needs to wait to be finished.
	StartTime *metav1.Time `json:"startTime,omitempty"`
}

// Plan is a list of actions that will be taken to update a deployment.
// Only 1 action is in progress at a time. The operator will wait for that
// action to be completely and then remove the action.
type Plan []Action
