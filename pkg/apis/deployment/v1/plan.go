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

package v1

import (
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/dchest/uniuri"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ActionType is a strongly typed name for a plan action item
type ActionType string

const (
	// ActionTypeAddMember causes a member to be added.
	ActionTypeAddMember ActionType = "AddMember"
	// ActionTypeRemoveMember causes a member to be removed.
	ActionTypeRemoveMember ActionType = "RemoveMember"
	// ActionTypeRecreateMember recreates member. Used when member is still owner of some shards.
	ActionTypeRecreateMember ActionType = "RecreateMember"
	// ActionTypeCleanOutMember causes a member to be cleaned out (dbserver only).
	ActionTypeCleanOutMember ActionType = "CleanOutMember"
	// ActionTypeShutdownMember causes a member to be shutdown and removed from the cluster.
	ActionTypeShutdownMember ActionType = "ShutdownMember"
	// ActionTypeRotateMember causes a member to be shutdown and have it's pod removed.
	ActionTypeRotateMember ActionType = "RotateMember"
	// ActionTypeUpgradeMember causes a member to be shutdown and have it's pod removed, restarted with AutoUpgrade option, waited until termination and the restarted again.
	ActionTypeUpgradeMember ActionType = "UpgradeMember"
	// ActionTypeWaitForMemberUp causes the plan to wait until the member is considered "up".
	ActionTypeWaitForMemberUp ActionType = "WaitForMemberUp"
	// ActionTypeRenewTLSCertificate causes the TLS certificate of a member to be renewed.
	ActionTypeRenewTLSCertificate ActionType = "RenewTLSCertificate"
	// ActionTypeRenewTLSCACertificate causes the TLS CA certificate of the entire deployment to be renewed.
	ActionTypeRenewTLSCACertificate ActionType = "RenewTLSCACertificate"
	// ActionTypeSetCurrentImage causes status.CurrentImage to be updated to the image given in the action.
	ActionTypeSetCurrentImage ActionType = "SetCurrentImage"
	// ActionTypeDisableClusterScaling turns off scaling DBservers and coordinators
	ActionTypeDisableClusterScaling ActionType = "ScalingDisabled"
	// ActionTypeEnableClusterScaling turns on scaling DBservers and coordinators
	ActionTypeEnableClusterScaling ActionType = "ScalingEnabled"
)

const (
	// MemberIDPreviousAction is used for Action.MemberID when the MemberID
	// should be derived from the previous action.
	MemberIDPreviousAction = "@previous"
)

// Action represents a single action to be taken to update a deployment.
type Action struct {
	// ID of this action (unique for every action)
	ID string `json:"id"`
	// Type of action.
	Type ActionType `json:"type"`
	// ID reference of the member involved in this action (if any)
	MemberID string `json:"memberID,omitempty"`
	// Group involved in this action
	Group ServerGroup `json:"group,omitempty"`
	// CreationTime is set the when the action is created.
	CreationTime metav1.Time `json:"creationTime"`
	// StartTime is set the when the action has been started, but needs to wait to be finished.
	StartTime *metav1.Time `json:"startTime,omitempty"`
	// Reason for this action
	Reason string `json:"reason,omitempty"`
	// Image used in can of a SetCurrentImage action.
	Image string `json:"image,omitempty"`
}

// Equal compares two Actions
func (a Action) Equal(other Action) bool {
	return a.ID == other.ID &&
		a.Type == other.Type &&
		a.MemberID == other.MemberID &&
		a.Group == other.Group &&
		util.TimeCompareEqual(a.CreationTime, other.CreationTime) &&
		util.TimeCompareEqualPointer(a.StartTime, other.StartTime) &&
		a.Reason == other.Reason &&
		a.Image == other.Image
}

// NewAction instantiates a new Action.
func NewAction(actionType ActionType, group ServerGroup, memberID string, reason ...string) Action {
	a := Action{
		ID:           uniuri.New(),
		Type:         actionType,
		MemberID:     memberID,
		Group:        group,
		CreationTime: metav1.Now(),
	}
	if len(reason) != 0 {
		a.Reason = reason[0]
	}
	return a
}

// SetImage sets the Image field to the given value and returns the modified
// action.
func (a Action) SetImage(image string) Action {
	a.Image = image
	return a
}

// Plan is a list of actions that will be taken to update a deployment.
// Only 1 action is in progress at a time. The operator will wait for that
// action to be completely and then remove the action.
type Plan []Action

// Equal compares two Plan
func (p Plan) Equal(other Plan) bool {
	// For plan the order is relevant!
	if len(p) != len(other) {
		return false
	}

	for i := 0; i < len(p); i++ {
		if !p[i].Equal(other[i]) {
			return false
		}
	}

	return true
}

// IsEmpty checks if plan is empty
func (p Plan) IsEmpty() bool {
	return len(p) == 0
}
