//
// DISCLAIMER
//
// Copyright 2026 ArangoDB GmbH, Cologne, Germany
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

package resources

import (
	"testing"

	"github.com/stretchr/testify/require"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func Test_DefaultResourceList(t *testing.T) {
	l := DefaultResourceList(core.ResourceList{
		core.ResourceCPU:              resource.MustParse("2"),
		core.ResourceEphemeralStorage: resource.MustParse("4Gi"),
	}, resource.MustParse("1Gi"), core.ResourceCPU, core.ResourceMemory)

	require.Contains(t, l, core.ResourceCPU)
	require.Equal(t, resource.MustParse("2"), l[core.ResourceCPU])

	require.Contains(t, l, core.ResourceEphemeralStorage)
	require.Equal(t, resource.MustParse("4Gi"), l[core.ResourceEphemeralStorage])

	require.Contains(t, l, core.ResourceMemory)
	require.Equal(t, resource.MustParse("1Gi"), l[core.ResourceMemory])
}

func Test_ScaleQuantity(t *testing.T) {
	require.True(t, resource.MustParse("3").Equal(ScaleQuantity(resource.MustParse("4"), core.ResourceCPU, 0.75)))
	require.True(t, resource.MustParse("3Gi").Equal(ScaleQuantity(resource.MustParse("4Gi"), core.ResourceMemory, 0.75)))
	require.True(t, resource.MustParse("2Gi").Equal(ScaleQuantity(resource.MustParse("4Gi"), core.ResourceEphemeralStorage, 0.5)))
	require.True(t, resource.MustParse("0").Equal(ScaleQuantity(resource.MustParse("4Gi"), core.ResourcePods, 0.5)))
}

func Test_ScaleResources(t *testing.T) {
	in := core.ResourceRequirements{
		Limits: core.ResourceList{
			core.ResourceCPU:              resource.MustParse("4"),
			core.ResourceEphemeralStorage: resource.MustParse("4Gi"),
			core.ResourceMemory:           resource.MustParse("2Gi"),
			core.ResourcePods:             resource.MustParse("3"),
		},
		Requests: core.ResourceList{
			core.ResourceCPU:              resource.MustParse("8"),
			core.ResourceEphemeralStorage: resource.MustParse("8Gi"),
			core.ResourceMemory:           resource.MustParse("4Gi"),
			core.ResourcePods:             resource.MustParse("6"),
		},
	}

	t.Run("Scale 1", func(t *testing.T) {
		generated := ScaleResources(in, 1.0)

		require.True(t, resource.MustParse("4").Equal(generated.Limits[core.ResourceCPU]))
		require.True(t, resource.MustParse("4Gi").Equal(generated.Limits[core.ResourceEphemeralStorage]))
		require.True(t, resource.MustParse("2Gi").Equal(generated.Limits[core.ResourceMemory]))
		require.True(t, resource.MustParse("3").Equal(generated.Limits[core.ResourcePods]))

		require.True(t, resource.MustParse("8").Equal(generated.Requests[core.ResourceCPU]))
		require.True(t, resource.MustParse("8Gi").Equal(generated.Requests[core.ResourceEphemeralStorage]))
		require.True(t, resource.MustParse("4Gi").Equal(generated.Requests[core.ResourceMemory]))
		require.True(t, resource.MustParse("6").Equal(generated.Requests[core.ResourcePods]))
	})

	t.Run("Scale 0.5", func(t *testing.T) {
		generated := ScaleResources(in, 0.5)

		require.True(t, resource.MustParse("2").Equal(generated.Limits[core.ResourceCPU]))
		require.True(t, resource.MustParse("2Gi").Equal(generated.Limits[core.ResourceEphemeralStorage]))
		require.True(t, resource.MustParse("1Gi").Equal(generated.Limits[core.ResourceMemory]))
		require.True(t, resource.MustParse("0").Equal(generated.Limits[core.ResourcePods]))

		require.True(t, resource.MustParse("4").Equal(generated.Requests[core.ResourceCPU]))
		require.True(t, resource.MustParse("4Gi").Equal(generated.Requests[core.ResourceEphemeralStorage]))
		require.True(t, resource.MustParse("2Gi").Equal(generated.Requests[core.ResourceMemory]))
		require.True(t, resource.MustParse("0").Equal(generated.Requests[core.ResourcePods]))
	})
}
