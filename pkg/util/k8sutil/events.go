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
	"os"
	"strings"
	"time"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/util/constants"
)

// APIObject helps to abstract an object from our custom API.
type APIObject interface {
	metav1.Object
	// AsOwner creates an OwnerReference for the given deployment
	AsOwner() metav1.OwnerReference
}

// NewMemberAddEvent creates an event indicating that a member was added.
func NewMemberAddEvent(memberName, role string, apiObject APIObject) *v1.Event {
	event := newDeploymentEvent(apiObject)
	event.Type = v1.EventTypeNormal
	event.Reason = fmt.Sprintf("New %s Added", strings.Title(role))
	event.Message = fmt.Sprintf("New %s %s added to deployment", role, memberName)
	return event
}

// NewMemberRemoveEvent creates an event indicating that an existing member was removed.
func NewMemberRemoveEvent(memberName, role string, apiObject APIObject) *v1.Event {
	event := newDeploymentEvent(apiObject)
	event.Type = v1.EventTypeNormal
	event.Reason = fmt.Sprintf("%s Removed", strings.Title(role))
	event.Message = fmt.Sprintf("Existing %s %s removed from the deployment", role, memberName)
	return event
}

// NewPodGoneEvent creates an event indicating that a pod is missing
func NewPodGoneEvent(podName, role string, apiObject APIObject) *v1.Event {
	event := newDeploymentEvent(apiObject)
	event.Type = v1.EventTypeNormal
	event.Reason = fmt.Sprintf("Pod Of %s Gone", strings.Title(role))
	event.Message = fmt.Sprintf("Pod %s of member %s is gone", podName, role)
	return event
}

// NewImmutableFieldEvent creates an event indicating that an attempt was made to change a field
// that is immutable.
func NewImmutableFieldEvent(fieldName string, apiObject APIObject) *v1.Event {
	event := newDeploymentEvent(apiObject)
	event.Type = v1.EventTypeNormal
	event.Reason = "Immutable Field Change"
	event.Message = fmt.Sprintf("Changing field %s is not possible. It has been reset to its original value.", fieldName)
	return event
}

// NewPodsSchedulingFailureEvent creates an event indicating that one of more cannot be scheduled.
func NewPodsSchedulingFailureEvent(unscheduledPodNames []string, apiObject APIObject) *v1.Event {
	event := newDeploymentEvent(apiObject)
	event.Type = v1.EventTypeNormal
	event.Reason = "Pods Scheduling Failure"
	event.Message = fmt.Sprintf("One or more pods are not scheduled in time. Pods: %v", unscheduledPodNames)
	return event
}

// NewPodsSchedulingResolvedEvent creates an event indicating that an earlier problem with
// pod scheduling has been resolved.
func NewPodsSchedulingResolvedEvent(apiObject APIObject) *v1.Event {
	event := newDeploymentEvent(apiObject)
	event.Type = v1.EventTypeNormal
	event.Reason = "Pods Scheduling Resolved"
	event.Message = "All pods have been scheduled"
	return event
}

// NewSecretsChangedEvent creates an event indicating that one of more secrets have changed.
func NewSecretsChangedEvent(changedSecretNames []string, apiObject APIObject) *v1.Event {
	event := newDeploymentEvent(apiObject)
	event.Type = v1.EventTypeNormal
	event.Reason = "Secrets changed"
	event.Message = fmt.Sprintf("Found %d changed secrets. You must revert them before the operator can continue. Secrets: %v", len(changedSecretNames), changedSecretNames)
	return event
}

// NewSecretsRestoredEvent creates an event indicating that all secrets have been restored
// to their original values.
func NewSecretsRestoredEvent(apiObject APIObject) *v1.Event {
	event := newDeploymentEvent(apiObject)
	event.Type = v1.EventTypeNormal
	event.Reason = "Secrets restored"
	event.Message = "All secrets have been restored to their original value"
	return event
}

// NewErrorEvent creates an even of type error.
func NewErrorEvent(reason string, err error, apiObject APIObject) *v1.Event {
	event := newDeploymentEvent(apiObject)
	event.Type = v1.EventTypeWarning
	event.Reason = strings.Title(reason)
	event.Message = err.Error()
	return event
}

// newDeploymentEvent creates a new event for the given api object & owner.
func newDeploymentEvent(apiObject APIObject) *v1.Event {
	t := time.Now()
	owner := apiObject.AsOwner()
	return &v1.Event{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: apiObject.GetName() + "-",
			Namespace:    apiObject.GetNamespace(),
		},
		InvolvedObject: v1.ObjectReference{
			APIVersion:      owner.APIVersion,
			Kind:            owner.Kind,
			Name:            owner.Name,
			Namespace:       apiObject.GetNamespace(),
			UID:             owner.UID,
			ResourceVersion: apiObject.GetResourceVersion(),
		},
		Source: v1.EventSource{
			Component: os.Getenv(constants.EnvOperatorPodName),
		},
		// Each deployment event is unique so it should not be collapsed with other events
		FirstTimestamp: metav1.Time{Time: t},
		LastTimestamp:  metav1.Time{Time: t},
		Count:          int32(1),
	}
}
