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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/kubernetes/fake"
)

func Test_Event_Handler(t *testing.T) {
	// Arrange
	c := fake.NewSimpleClientset()

	recorder := NewEventRecorder("mock", c)

	group := string(uuid.NewUUID())
	version := "v1"
	kind := string(uuid.NewUUID())

	name := string(uuid.NewUUID())
	namespace := string(uuid.NewUUID())

	reason := string(uuid.NewUUID())
	message := string(uuid.NewUUID())

	instance := recorder.NewInstance(group, version, kind)

	p := &core.Pod{
		ObjectMeta: meta.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	cases := map[string]func(object meta.Object, reason, format string, a ...interface{}){
		core.EventTypeNormal:  instance.Normal,
		core.EventTypeWarning: instance.Warning,
	}

	// Act
	for eventType, call := range cases {
		t.Run(eventType, func(t *testing.T) {
			call(p, reason, message)

			// Assert
			events, err := c.CoreV1().Events(namespace).List(context.Background(), meta.ListOptions{})
			require.NoError(t, err)
			require.Len(t, events.Items, 1)

			event := events.Items[0]
			assert.Equal(t, eventType, event.Type)
			assert.Equal(t, reason, event.Reason)
			assert.Equal(t, message, event.Message)
			assert.Equal(t, "mock", event.Source.Component)

			assert.Equal(t, fmt.Sprintf("%s/%s", group, version), event.InvolvedObject.APIVersion)
			assert.Equal(t, kind, event.InvolvedObject.Kind)
			assert.Equal(t, namespace, event.InvolvedObject.Namespace)
			assert.Equal(t, name, event.InvolvedObject.Name)

			require.NoError(t, c.CoreV1().Events(namespace).Delete(context.Background(), event.Name, meta.DeleteOptions{}))
		})
	}
}
