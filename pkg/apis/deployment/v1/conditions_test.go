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

package v1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConditionListIsTrue(t *testing.T) {
	assert.False(t, ConditionList{}.IsTrue(ConditionTypeReady))

	cl := ConditionList{}
	cl.Update(ConditionTypeReady, true, "test", "msg")
	assert.True(t, cl.IsTrue(ConditionTypeReady))
	assert.False(t, cl.IsTrue(ConditionTypeTerminated))

	cl.Update(ConditionTypeReady, false, "test", "msg")
	assert.False(t, cl.IsTrue(ConditionTypeReady))

	cl.Remove(ConditionTypeReady)
	assert.False(t, cl.IsTrue(ConditionTypeReady))
	assert.Equal(t, 0, len(cl))
}

func TestConditionListGet(t *testing.T) {
	conv := func(c Condition, b bool) []interface{} {
		return []interface{}{c, b}
	}

	cl := ConditionList{}
	assert.EqualValues(t, conv(Condition{}, false), conv(cl.Get(ConditionTypeReady)))
	cl.Update(ConditionTypeReady, false, "test", "msg")
	assert.EqualValues(t, conv(cl[0], true), conv(cl.Get(ConditionTypeReady)))
}

func TestConditionListUpdate(t *testing.T) {
	cl := ConditionList{}
	assert.Equal(t, 0, len(cl))

	assert.True(t, cl.Update(ConditionTypeReady, true, "test", "msg"))
	assert.True(t, cl.IsTrue(ConditionTypeReady))
	assert.Equal(t, 1, len(cl))

	assert.False(t, cl.Update(ConditionTypeReady, true, "test", "msg"))
	assert.True(t, cl.IsTrue(ConditionTypeReady))
	assert.Equal(t, 1, len(cl))

	assert.True(t, cl.Update(ConditionTypeReady, false, "test", "msg"))
	assert.False(t, cl.IsTrue(ConditionTypeReady))
	assert.Equal(t, 1, len(cl))

	assert.True(t, cl.Update(ConditionTypeReady, false, "test2", "msg"))
	assert.False(t, cl.IsTrue(ConditionTypeReady))
	assert.Equal(t, 1, len(cl))

	assert.True(t, cl.Update(ConditionTypeReady, false, "test2", "msg2"))
	assert.False(t, cl.IsTrue(ConditionTypeReady))
	assert.Equal(t, 1, len(cl))
}

func TestConditionListRemove(t *testing.T) {
	cl := ConditionList{}
	assert.Equal(t, 0, len(cl))

	cl.Update(ConditionTypeReady, true, "test", "msg")
	cl.Update(ConditionTypeTerminated, false, "test", "msg")
	assert.Equal(t, 2, len(cl))

	assert.True(t, cl.Remove(ConditionTypeReady))
	assert.Equal(t, 1, len(cl))

	assert.False(t, cl.Remove(ConditionTypeReady))
	assert.Equal(t, 1, len(cl))

	assert.True(t, cl.Remove(ConditionTypeTerminated))
	assert.Equal(t, 0, len(cl))
}
