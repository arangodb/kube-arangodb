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

func TestConditionListIsTrue(t *testing.T) {
	assert.False(t, ConditionList{}.IsTrue(ConditionTypeConfigured))

	cl := ConditionList{}
	cl.Update(ConditionTypeConfigured, true, "test", "msg")
	assert.True(t, cl.IsTrue(ConditionTypeConfigured))
	//assert.False(t, cl.IsTrue(ConditionTypeTerminated))

	cl.Update(ConditionTypeConfigured, false, "test", "msg")
	assert.False(t, cl.IsTrue(ConditionTypeConfigured))

	cl.Remove(ConditionTypeConfigured)
	assert.False(t, cl.IsTrue(ConditionTypeConfigured))
	assert.Equal(t, 0, len(cl))
}

func TestConditionListGet(t *testing.T) {
	conv := func(c Condition, b bool) []interface{} {
		return []interface{}{c, b}
	}

	cl := ConditionList{}
	assert.EqualValues(t, conv(Condition{}, false), conv(cl.Get(ConditionTypeConfigured)))
	cl.Update(ConditionTypeConfigured, false, "test", "msg")
	assert.EqualValues(t, conv(cl[0], true), conv(cl.Get(ConditionTypeConfigured)))
}

func TestConditionListUpdate(t *testing.T) {
	cl := ConditionList{}
	assert.Equal(t, 0, len(cl))

	assert.True(t, cl.Update(ConditionTypeConfigured, true, "test", "msg"))
	assert.True(t, cl.IsTrue(ConditionTypeConfigured))
	assert.Equal(t, 1, len(cl))

	assert.False(t, cl.Update(ConditionTypeConfigured, true, "test", "msg"))
	assert.True(t, cl.IsTrue(ConditionTypeConfigured))
	assert.Equal(t, 1, len(cl))

	assert.True(t, cl.Update(ConditionTypeConfigured, false, "test", "msg"))
	assert.False(t, cl.IsTrue(ConditionTypeConfigured))
	assert.Equal(t, 1, len(cl))

	assert.True(t, cl.Update(ConditionTypeConfigured, false, "test2", "msg"))
	assert.False(t, cl.IsTrue(ConditionTypeConfigured))
	assert.Equal(t, 1, len(cl))

	assert.True(t, cl.Update(ConditionTypeConfigured, false, "test2", "msg2"))
	assert.False(t, cl.IsTrue(ConditionTypeConfigured))
	assert.Equal(t, 1, len(cl))
}

func TestConditionListRemove(t *testing.T) {
	cl := ConditionList{}
	assert.Equal(t, 0, len(cl))

	cl.Update(ConditionTypeConfigured, true, "test", "msg")
	assert.Equal(t, 1, len(cl))

	assert.True(t, cl.Remove(ConditionTypeConfigured))
	assert.Equal(t, 0, len(cl))

	assert.False(t, cl.Remove(ConditionTypeConfigured))
	assert.Equal(t, 0, len(cl))
}
