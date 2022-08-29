//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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

package event

import (
	"context"
	"fmt"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/kubernetes"
)

// NewEventRecorder creates new event recorder
func NewEventRecorder(name string, kubeClientSet kubernetes.Interface) Recorder {
	return &eventRecorder{
		kubeClientSet: kubeClientSet,
		name:          name,
	}
}

// Recorder event factory for kubernetes
type Recorder interface {
	NewInstance(group, version, kind string) RecorderInstance

	event(group, version, kind string, object meta.Object, eventType, reason, message string)
}

type eventRecorder struct {
	name          string
	kubeClientSet kubernetes.Interface
}

func (e *eventRecorder) newEvent(group, version, kind string, object meta.Object, eventType, reason, message string) *core.Event {
	return &core.Event{
		InvolvedObject: e.newObjectReference(group, version, kind, object),

		ReportingController: e.name,

		Type:    eventType,
		Reason:  reason,
		Message: message,

		ObjectMeta: meta.ObjectMeta{
			Namespace: object.GetNamespace(),
			Name:      string(uuid.NewUUID()),
		},

		FirstTimestamp: meta.Now(),
		LastTimestamp:  meta.Now(),

		Source: core.EventSource{
			Component: e.name,
		},
	}
}

func (e *eventRecorder) newObjectReference(group, version, kind string, object meta.Object) core.ObjectReference {
	return core.ObjectReference{
		UID:        object.GetUID(),
		APIVersion: fmt.Sprintf("%s/%s", group, version),
		Kind:       kind,
		Name:       object.GetName(),
		Namespace:  object.GetNamespace(),
	}
}

func (e *eventRecorder) event(group, version, kind string, object meta.Object, eventType, reason, message string) {
	_, err := e.kubeClientSet.CoreV1().Events(object.GetNamespace()).Create(context.Background(), e.newEvent(group, version, kind, object, eventType, reason, message), meta.CreateOptions{})
	if err != nil {
		logger.Err(err).
			Str("APIVersion", fmt.Sprintf("%s/%s", group, version)).
			Str("Kind", kind).
			Str("Object", fmt.Sprintf("%s/%s", object.GetNamespace(), object.GetName())).
			Warn("Unable to send event")
		return
	}

	logger.
		Str("APIVersion", fmt.Sprintf("%s/%s", group, version)).
		Str("Kind", kind).
		Str("Object", fmt.Sprintf("%s/%s", object.GetNamespace(), object.GetName())).
		Info("Event send %s - %s - %s", eventType, reason, message)
}

func (e *eventRecorder) NewInstance(group, version, kind string) RecorderInstance {
	return &eventRecorderInstance{
		group:   group,
		version: version,
		kind:    kind,

		eventRecorder: e,
	}
}

// RecorderInstance represents instance of event recorder for specific kubernetes type
type RecorderInstance interface {
	Event(object meta.Object, eventType, reason, format string, a ...interface{})

	Warning(object meta.Object, reason, format string, a ...interface{})
	Normal(object meta.Object, reason, format string, a ...interface{})
}

type eventRecorderInstance struct {
	group, version, kind string

	eventRecorder Recorder
}

func (e *eventRecorderInstance) Warning(object meta.Object, reason, format string, a ...interface{}) {
	e.Event(object, core.EventTypeWarning, reason, format, a...)
}

func (e *eventRecorderInstance) Normal(object meta.Object, reason, format string, a ...interface{}) {
	e.Event(object, core.EventTypeNormal, reason, format, a...)
}

func (e *eventRecorderInstance) Event(object meta.Object, eventType, reason, format string, a ...interface{}) {
	e.eventRecorder.event(e.group, e.version, e.kind, object, eventType, reason, fmt.Sprintf(format, a...))
}
