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

package k8sutil

import (
	"fmt"
	"strings"

	driver "github.com/arangodb/go-driver"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// Event is used to create events using an EventRecorder.
type Event struct {
	InvolvedObject runtime.Object
	Type           string
	Reason         string
	Message        string
}

// APIObject helps to abstract an object from our custom API.
type APIObject interface {
	runtime.Object
	metav1.Object
	// AsOwner creates an OwnerReference for the given deployment
	AsOwner() metav1.OwnerReference
}

// NewMemberAddEvent creates an event indicating that a member was added.
func NewMemberAddEvent(memberName, role string, apiObject APIObject) *Event {
	event := newDeploymentEvent(apiObject)
	event.Type = v1.EventTypeNormal
	event.Reason = fmt.Sprintf("New %s Added", strings.Title(role))
	event.Message = fmt.Sprintf("New %s %s added to deployment", role, memberName)
	return event
}

// NewMemberRemoveEvent creates an event indicating that an existing member was removed.
func NewMemberRemoveEvent(memberName, role string, apiObject APIObject) *Event {
	event := newDeploymentEvent(apiObject)
	event.Type = v1.EventTypeNormal
	event.Reason = fmt.Sprintf("%s Removed", strings.Title(role))
	event.Message = fmt.Sprintf("Existing %s %s removed from the deployment", role, memberName)
	return event
}

// NewPodCreatedEvent creates an event indicating that a pod has been created
func NewPodCreatedEvent(podName, role string, apiObject APIObject) *Event {
	event := newDeploymentEvent(apiObject)
	event.Type = v1.EventTypeNormal
	event.Reason = fmt.Sprintf("Pod Of %s Created", strings.Title(role))
	event.Message = fmt.Sprintf("Pod %s of member %s is created", podName, role)
	return event
}

// NewPodGoneEvent creates an event indicating that a pod is missing
func NewPodGoneEvent(podName, role string, apiObject APIObject) *Event {
	event := newDeploymentEvent(apiObject)
	event.Type = v1.EventTypeNormal
	event.Reason = fmt.Sprintf("Pod Of %s Gone", strings.Title(role))
	event.Message = fmt.Sprintf("Pod %s of member %s is gone", podName, role)
	return event
}

// NewImmutableFieldEvent creates an event indicating that an attempt was made to change a field
// that is immutable.
func NewImmutableFieldEvent(fieldName string, apiObject APIObject) *Event {
	event := newDeploymentEvent(apiObject)
	event.Type = v1.EventTypeNormal
	event.Reason = "Immutable Field Change"
	event.Message = fmt.Sprintf("Changing field %s is not possible. It has been reset to its original value.", fieldName)
	return event
}

// NewPodsSchedulingFailureEvent creates an event indicating that one of more cannot be scheduled.
func NewPodsSchedulingFailureEvent(unscheduledPodNames []string, apiObject APIObject) *Event {
	event := newDeploymentEvent(apiObject)
	event.Type = v1.EventTypeNormal
	event.Reason = "Pods Scheduling Failure"
	event.Message = fmt.Sprintf("One or more pods are not scheduled in time. Pods: %v", unscheduledPodNames)
	return event
}

// NewPodsSchedulingResolvedEvent creates an event indicating that an earlier problem with
// pod scheduling has been resolved.
func NewPodsSchedulingResolvedEvent(apiObject APIObject) *Event {
	event := newDeploymentEvent(apiObject)
	event.Type = v1.EventTypeNormal
	event.Reason = "Pods Scheduling Resolved"
	event.Message = "All pods have been scheduled"
	return event
}

// NewSecretsChangedEvent creates an event indicating that one of more secrets have changed.
func NewSecretsChangedEvent(changedSecretNames []string, apiObject APIObject) *Event {
	event := newDeploymentEvent(apiObject)
	event.Type = v1.EventTypeNormal
	event.Reason = "Secrets changed"
	event.Message = fmt.Sprintf("Found %d changed secrets. You must revert them before the operator can continue. Secrets: %v", len(changedSecretNames), changedSecretNames)
	return event
}

// NewSecretsRestoredEvent creates an event indicating that all secrets have been restored
// to their original values.
func NewSecretsRestoredEvent(apiObject APIObject) *Event {
	event := newDeploymentEvent(apiObject)
	event.Type = v1.EventTypeNormal
	event.Reason = "Secrets restored"
	event.Message = "All secrets have been restored to their original value"
	return event
}

// NewAccessPackageCreatedEvent creates an event indicating that a secret containing an access package
// has been created.
func NewAccessPackageCreatedEvent(apiObject APIObject, apSecretName string) *Event {
	event := newDeploymentEvent(apiObject)
	event.Type = v1.EventTypeNormal
	event.Reason = "Access package created"
	event.Message = fmt.Sprintf("An access package named %s has been created", apSecretName)
	return event
}

// NewAccessPackageDeletedEvent creates an event indicating that a secret containing an access package
// has been deleted.
func NewAccessPackageDeletedEvent(apiObject APIObject, apSecretName string) *Event {
	event := newDeploymentEvent(apiObject)
	event.Type = v1.EventTypeNormal
	event.Reason = "Access package deleted"
	event.Message = fmt.Sprintf("An access package named %s has been deleted", apSecretName)
	return event
}

// NewPlanTimeoutEvent creates an event indicating that an item on a reconciliation plan did not
// finish before its deadline.
func NewPlanTimeoutEvent(apiObject APIObject, itemType, memberID, role string) *Event {
	event := newDeploymentEvent(apiObject)
	event.Type = v1.EventTypeNormal
	event.Reason = "Reconciliation Plan Timeout"
	event.Message = fmt.Sprintf("An plan item of type %s or member %s with role %s did not finish in time", itemType, memberID, role)
	return event
}

// NewPlanAbortedEvent creates an event indicating that an item on a reconciliation plan wants to abort
// the entire plan.
func NewPlanAbortedEvent(apiObject APIObject, itemType, memberID, role string) *Event {
	event := newDeploymentEvent(apiObject)
	event.Type = v1.EventTypeNormal
	event.Reason = "Reconciliation Plan Aborted"
	event.Message = fmt.Sprintf("An plan item of type %s or member %s with role %s wants to abort the plan", itemType, memberID, role)
	return event
}

// NewCannotChangeStorageClassEvent creates an event indicating that an item would need to use a different StorageClass,
// but this is not possible for the given reason.
func NewCannotChangeStorageClassEvent(apiObject APIObject, memberID, role, subReason string) *Event {
	event := newDeploymentEvent(apiObject)
	event.Type = v1.EventTypeNormal
	event.Reason = fmt.Sprintf("%s Member StorageClass Cannot Change", strings.Title(role))
	event.Message = fmt.Sprintf("Member %s with role %s should use a different StorageClass, but is cannot because: %s", memberID, role, subReason)
	return event
}

// NewDowntimeNotAllowedEvent creates an event indicating that an operation cannot be executed because downtime
// is currently not allowed.
func NewDowntimeNotAllowedEvent(apiObject APIObject, operation string) *Event {
	event := newDeploymentEvent(apiObject)
	event.Type = v1.EventTypeNormal
	event.Reason = "Downtime Operation Postponed"
	event.Message = fmt.Sprintf("The '%s' operation is postponed because downtime it not allowed. Set `spec.downtimeAllowed` to true to execute this operation", operation)
	return event
}

// NewUpgradeNotAllowedEvent creates an event indicating that an upgrade (or downgrade) is not allowed.
func NewUpgradeNotAllowedEvent(apiObject APIObject, fromVersion, toVersion driver.Version) *Event {
	event := newDeploymentEvent(apiObject)
	event.Type = v1.EventTypeNormal
	if fromVersion.CompareTo(toVersion) < 0 {
		event.Reason = "Upgrade not allowed"
		event.Message = fmt.Sprintf("Upgrading ArangoDB from version %s to %s is not allowed", fromVersion, toVersion)
	} else {
		event.Reason = "Downgrade not allowed"
		event.Message = fmt.Sprintf("Downgrading ArangoDB from version %s to %s is not allowed", fromVersion, toVersion)
	}
	return event
}

// NewErrorEvent creates an even of type error.
func NewErrorEvent(reason string, err error, apiObject APIObject) *Event {
	event := newDeploymentEvent(apiObject)
	event.Type = v1.EventTypeWarning
	event.Reason = strings.Title(reason)
	event.Message = err.Error()
	return event
}

// newDeploymentEvent creates a new event for the given api object & owner.
func newDeploymentEvent(apiObject runtime.Object) *Event {
	return &Event{
		InvolvedObject: apiObject,
	}
}
