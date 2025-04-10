//
// DISCLAIMER
//
// Copyright 2016-2024 ArangoDB GmbH, Cologne, Germany
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

package metric_descriptions

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_ArangodbOperatorMembersUnexpectedContainerExitCodes_Descriptor(t *testing.T) {
	ArangodbOperatorMembersUnexpectedContainerExitCodes()
}

func Test_ArangodbOperatorMembersUnexpectedContainerExitCodes_Factory(t *testing.T) {
	global := NewArangodbOperatorMembersUnexpectedContainerExitCodesCounterFactory()

	object1 := ArangodbOperatorMembersUnexpectedContainerExitCodesInput{
		Namespace:     "1",
		Name:          "1",
		Member:        "1",
		Container:     "1",
		ContainerType: "1",
		Code:          "1",
		Reason:        "1",
	}

	object2 := ArangodbOperatorMembersUnexpectedContainerExitCodesInput{
		Namespace:     "2",
		Name:          "2",
		Member:        "2",
		Container:     "2",
		ContainerType: "2",
		Code:          "2",
		Reason:        "2",
	}

	t.Run("List", func(t *testing.T) {
		require.Len(t, global.Items(), 0)
	})

	t.Run("Precheck", func(t *testing.T) {
		require.EqualValues(t, 0, global.Get(object1))
		require.EqualValues(t, 0, global.Get(object2))
	})

	t.Run("Add", func(t *testing.T) {
		global.Add(object1, 10)

		require.EqualValues(t, 10, global.Get(object1))
		require.EqualValues(t, 0, global.Get(object2))
	})

	t.Run("List", func(t *testing.T) {
		require.Len(t, global.Items(), 1)
	})

	t.Run("Add", func(t *testing.T) {
		global.Add(object2, 3)

		require.EqualValues(t, 10, global.Get(object1))
		require.EqualValues(t, 3, global.Get(object2))
	})

	t.Run("List", func(t *testing.T) {
		require.Len(t, global.Items(), 2)
	})

	t.Run("Dec", func(t *testing.T) {
		global.Add(object1, -1)

		require.EqualValues(t, 9, global.Get(object1))
		require.EqualValues(t, 3, global.Get(object2))
	})

	t.Run("List", func(t *testing.T) {
		require.Len(t, global.Items(), 2)
	})

	t.Run("Remove", func(t *testing.T) {
		global.Remove(object1)

		require.EqualValues(t, 0, global.Get(object1))
		require.EqualValues(t, 3, global.Get(object2))
	})

	t.Run("List", func(t *testing.T) {
		require.Len(t, global.Items(), 1)
	})

	t.Run("Remove", func(t *testing.T) {
		global.Remove(object1)

		require.EqualValues(t, 0, global.Get(object1))
		require.EqualValues(t, 3, global.Get(object2))
	})

	t.Run("List", func(t *testing.T) {
		require.Len(t, global.Items(), 1)
	})

	t.Run("Remove", func(t *testing.T) {
		global.Remove(object2)

		require.EqualValues(t, 0, global.Get(object1))
		require.EqualValues(t, 0, global.Get(object2))
	})

	t.Run("List", func(t *testing.T) {
		require.Len(t, global.Items(), 0)
	})
}

func Test_ArangodbOperatorMembersUnexpectedContainerExitCodes_Factory_Counter(t *testing.T) {
	global := NewArangodbOperatorMembersUnexpectedContainerExitCodesCounterFactory()

	object1 := ArangodbOperatorMembersUnexpectedContainerExitCodesInput{
		Namespace:     "1",
		Name:          "1",
		Member:        "1",
		Container:     "1",
		ContainerType: "1",
		Code:          "1",
		Reason:        "1",
	}

	object2 := ArangodbOperatorMembersUnexpectedContainerExitCodesInput{
		Namespace:     "2",
		Name:          "2",
		Member:        "2",
		Container:     "2",
		ContainerType: "2",
		Code:          "2",
		Reason:        "2",
	}

	t.Run("List", func(t *testing.T) {
		require.Len(t, global.Items(), 0)
	})

	t.Run("Precheck", func(t *testing.T) {
		require.EqualValues(t, 0, global.Get(object1))
		require.EqualValues(t, 0, global.Get(object2))
	})

	t.Run("Add", func(t *testing.T) {
		global.Add(object1, 10)

		require.EqualValues(t, 10, global.Get(object1))
		require.EqualValues(t, 0, global.Get(object2))
	})

	t.Run("List", func(t *testing.T) {
		require.Len(t, global.Items(), 1)
	})

	t.Run("Inc", func(t *testing.T) {
		global.Inc(object1)
		global.Inc(object2)

		require.EqualValues(t, 11, global.Get(object1))
		require.EqualValues(t, 1, global.Get(object2))
	})

	t.Run("List", func(t *testing.T) {
		require.Len(t, global.Items(), 2)
	})
}
