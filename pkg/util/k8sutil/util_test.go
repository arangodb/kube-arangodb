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

package k8sutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestAddOwnerRefToObject tests AddOwnerRefToObject.
func TestAddOwnerRefToObject(t *testing.T) {
	p := &core.Pod{}
	AddOwnerRefToObject(p, nil)
	assert.Len(t, p.GetOwnerReferences(), 0)

	AddOwnerRefToObject(p, &meta.OwnerReference{})
	assert.Len(t, p.GetOwnerReferences(), 1)
}

// UpdateOwnerRefToObjectIfNeeded tests UpdateOwnerRefToObjectIfNeeded.
func TestUpdateOwnerRefToObjectIfNeeded(t *testing.T) {
	p := &core.Pod{}
	result := UpdateOwnerRefToObjectIfNeeded(p, nil)
	assert.Len(t, p.GetOwnerReferences(), 0)
	assert.False(t, result)

	result = UpdateOwnerRefToObjectIfNeeded(p, &meta.OwnerReference{})
	assert.Len(t, p.GetOwnerReferences(), 1)
	assert.True(t, result)

	result = UpdateOwnerRefToObjectIfNeeded(p, &meta.OwnerReference{})
	assert.Len(t, p.GetOwnerReferences(), 1)
	assert.False(t, result)
}

// TestLabelsForDeployment tests LabelsForDeployment.
func TestLabelsForDeployment(t *testing.T) {
	l := LabelsForDeployment("test", "role")
	assert.Len(t, l, 3)
	assert.Equal(t, "arangodb", l["app"])
	assert.Equal(t, "role", l["role"])
	assert.Equal(t, "test", l["arango_deployment"])

	l = LabelsForDeployment("test", "")
	assert.Len(t, l, 2)
	assert.Equal(t, "arangodb", l["app"])
	assert.Equal(t, "test", l["arango_deployment"])
}
