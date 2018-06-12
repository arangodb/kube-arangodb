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

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestMemberStatusList tests modifying a MemberStatusList.
func TestMemberStatusList(t *testing.T) {
	list := &MemberStatusList{}
	m1 := MemberStatus{ID: "m1"}
	m2 := MemberStatus{ID: "m2"}
	m3 := MemberStatus{ID: "m3"}
	assert.Equal(t, 0, len(*list))

	assert.NoError(t, list.Add(m1))
	assert.Equal(t, 1, len(*list))

	assert.NoError(t, list.Add(m2))
	assert.NoError(t, list.Add(m3))
	assert.Equal(t, 3, len(*list))

	assert.Error(t, list.Add(m2))
	assert.Equal(t, 3, len(*list))

	assert.NoError(t, list.RemoveByID(m3.ID))
	assert.Equal(t, 2, len(*list))
	assert.False(t, list.ContainsID(m3.ID))

	m2.PodName = "foo"
	assert.NoError(t, list.Update(m2))
	assert.Equal(t, 2, len(*list))
	assert.True(t, list.ContainsID(m2.ID))
	x, found := list.ElementByPodName("foo")
	assert.True(t, found)
	assert.Equal(t, "foo", x.PodName)
	assert.Equal(t, m2.ID, x.ID)
}
