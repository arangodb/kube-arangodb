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

package v2alpha1

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

	assert.NoError(t, list.add(m1))
	assert.Equal(t, 1, len(*list))

	assert.NoError(t, list.add(m2))
	assert.NoError(t, list.add(m3))
	assert.Equal(t, 3, len(*list))

	assert.Error(t, list.add(m2))
	assert.Equal(t, 3, len(*list))

	assert.NoError(t, list.removeByID(m3.ID))
	assert.Equal(t, 2, len(*list))
	assert.False(t, list.ContainsID(m3.ID))
	assert.Equal(t, m1.ID, (*list)[0].ID)
	assert.Equal(t, m2.ID, (*list)[1].ID)

	m2.Pod = &MemberPodStatus{Name: "foo"}
	assert.NoError(t, list.update(m2))
	assert.Equal(t, 2, len(*list))
	assert.True(t, list.ContainsID(m2.ID))
	x, found := list.ElementByPodName("foo")
	assert.True(t, found)
	assert.Equal(t, "foo", x.Pod.GetName())
	assert.Equal(t, m2.ID, x.ID)

	assert.NoError(t, list.add(m3))
	assert.Equal(t, 3, len(*list))
	assert.Equal(t, m1.ID, (*list)[0].ID)
	assert.Equal(t, m2.ID, (*list)[1].ID)
	assert.Equal(t, m3.ID, (*list)[2].ID)

	list2 := &MemberStatusList{m3, m2, m1}
	assert.True(t, list.Equal(*list2))
	assert.True(t, list2.Equal(*list))

	list3 := &MemberStatusList{m3, m1}
	assert.False(t, list.Equal(*list3))
	assert.False(t, list3.Equal(*list))

	list4 := MemberStatusList{m3, m2, m1}
	list4[1].Phase = "something-else"
	assert.False(t, list.Equal(list4))
	assert.False(t, list4.Equal(*list))

	m4 := MemberStatus{ID: "m4"}
	list5 := &MemberStatusList{m1, m2, m4}
	assert.False(t, list.Equal(*list5))
	assert.False(t, list5.Equal(*list))

	list6 := &MemberStatusList{m1, m2, m3, m4}
	assert.False(t, list.Equal(*list6))
	assert.False(t, list6.Equal(*list))
}
