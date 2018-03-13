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
// Author Jan Christoph Uhde
//

package test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

var apiObjectForTest = api.ArangoDeployment{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "Willy",
		Namespace: "Wonka",
	},
	Spec: api.DeploymentSpec{
		Mode: api.DeploymentModeCluster,
	},
}

func TestMemberAddEvent(t *testing.T) {
	event := k8sutil.NewMemberAddEvent("name", "role", &apiObjectForTest)
	assert.Equal(t, event.Reason, "New Role Added")
	assert.Equal(t, event.Message, "New role name added to deployment")
	assert.Equal(t, event.Type, v1.EventTypeNormal)
}

func TestMemberRemoveEvent(t *testing.T) {
	event := k8sutil.NewMemberRemoveEvent("name", "role", &apiObjectForTest)
	assert.Equal(t, event.Reason, "Role Removed")
	assert.Equal(t, event.Message, "Existing role name removed from the deployment")
	assert.Equal(t, event.Type, v1.EventTypeNormal)
}

func TestPodGoneEvent(t *testing.T) {
	event := k8sutil.NewPodGoneEvent("name", "role", &apiObjectForTest)
	assert.Equal(t, event.Reason, "Pod Of Role Gone")
	assert.Equal(t, event.Message, "Pod name of member role is gone")
	assert.Equal(t, event.Type, v1.EventTypeNormal)
}

func TestImmutableFieldEvent(t *testing.T) {
	event := k8sutil.NewImmutableFieldEvent("name", &apiObjectForTest)
	assert.Equal(t, event.Reason, "Immutable Field Change")
	assert.Equal(t, event.Message, "Changing field name is not possible. It has been reset to its original value.")
	assert.Equal(t, event.Type, v1.EventTypeNormal)
}

func TestErrorEvent(t *testing.T) {
	event := k8sutil.NewErrorEvent("reason", errors.New("something"), &apiObjectForTest)
	assert.Equal(t, event.Reason, "Reason")
	assert.Equal(t, event.Type, v1.EventTypeWarning)
}
