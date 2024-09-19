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

func Test_ArangodbResourcesDeploymentConfigMapDuration_Descriptor(t *testing.T) {
	ArangodbResourcesDeploymentConfigMapDuration()
}

func Test_ArangodbResourcesDeploymentConfigMapDuration_Factory(t *testing.T) {
	global := NewArangodbResourcesDeploymentConfigMapDurationGaugeFactory()

	object1 := ArangodbResourcesDeploymentConfigMapDurationInput{
		Deployment: "1",
	}

	object2 := ArangodbResourcesDeploymentConfigMapDurationInput{
		Deployment: "2",
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

func Test_ArangodbResourcesDeploymentConfigMapDuration_Factory_Gauge(t *testing.T) {
	global := NewArangodbResourcesDeploymentConfigMapDurationGaugeFactory()

	object1 := ArangodbResourcesDeploymentConfigMapDurationInput{
		Deployment: "1",
	}

	object2 := ArangodbResourcesDeploymentConfigMapDurationInput{
		Deployment: "2",
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

	t.Run("Set", func(t *testing.T) {
		global.Set(object1, 3)
		global.Set(object2, 1)

		require.EqualValues(t, 3, global.Get(object1))
		require.EqualValues(t, 1, global.Get(object2))
	})

	t.Run("List", func(t *testing.T) {
		require.Len(t, global.Items(), 2)
	})
}
