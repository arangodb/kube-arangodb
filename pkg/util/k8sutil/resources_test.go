//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

func Test_ExtractPodAcceptedResourceRequirement(t *testing.T) {
	v, err := resource.ParseQuantity("1Gi")
	require.NoError(t, err)

	t.Run("Filter Storage", func(t *testing.T) {
		in := core.ResourceRequirements{
			Limits: map[core.ResourceName]resource.Quantity{
				core.ResourceCPU:     v,
				core.ResourceStorage: v,
			},
		}
		require.Len(t, in.Limits, 2)
		require.Len(t, in.Requests, 0)

		out := ExtractPodAcceptedResourceRequirement(in)
		require.Len(t, out.Limits, 1)
		require.Contains(t, out.Limits, core.ResourceCPU)
		require.NotContains(t, out.Limits, core.ResourceStorage)
		require.Len(t, out.Requests, 0)
	})

	t.Run("Ensure that all required Resources are filtered", func(t *testing.T) {
		resources := map[core.ResourceName]bool{
			core.ResourceCPU:              true,
			core.ResourceMemory:           true,
			core.ResourceStorage:          false,
			core.ResourceEphemeralStorage: true,
		}

		in := core.ResourceRequirements{
			Limits:   core.ResourceList{},
			Requests: core.ResourceList{},
		}

		for k := range resources {
			in.Limits[k] = v
			in.Requests[k] = v
		}

		out := ExtractPodAcceptedResourceRequirement(in)

		for k, v := range resources {
			t.Run(fmt.Sprintf("Resource %s should be %s", k, util.BoolSwitch(v, "present", "missing")), func(t *testing.T) {
				require.Contains(t, in.Requests, k)
				require.Contains(t, in.Limits, k)

				if v {
					require.Contains(t, out.Requests, k)
					require.Contains(t, out.Limits, k)
				} else {
					require.NotContains(t, out.Requests, k)
					require.NotContains(t, out.Limits, k)
				}
			})
		}
	})
}
func Test_ApplyContainerResourceRequirements(t *testing.T) {
	v1, err := resource.ParseQuantity("1Gi")
	require.NoError(t, err)

	v2, err := resource.ParseQuantity("4Gi")
	require.NoError(t, err)

	var container core.Container

	t.Run("Ensure limits are copied", func(t *testing.T) {
		ApplyContainerResourceRequirements(&container, core.ResourceRequirements{
			Limits: core.ResourceList{
				core.ResourceMemory: v1,
			},
			Requests: core.ResourceList{
				core.ResourceMemory: v1,
			},
		})

		require.Len(t, container.Resources.Requests, 1)
		require.Contains(t, container.Resources.Requests, core.ResourceMemory)
		require.Equal(t, v1, container.Resources.Requests[core.ResourceMemory])

		require.Len(t, container.Resources.Limits, 1)
		require.Contains(t, container.Resources.Limits, core.ResourceMemory)
		require.Equal(t, v1, container.Resources.Limits[core.ResourceMemory])
	})

	t.Run("Ensure limits are not overridden", func(t *testing.T) {
		ApplyContainerResourceRequirements(&container, core.ResourceRequirements{
			Limits: core.ResourceList{
				core.ResourceMemory: v2,
			},
			Requests: core.ResourceList{
				core.ResourceMemory: v2,
			},
		})

		require.Len(t, container.Resources.Requests, 1)
		require.Contains(t, container.Resources.Requests, core.ResourceMemory)
		require.Equal(t, v1, container.Resources.Requests[core.ResourceMemory])

		require.Len(t, container.Resources.Limits, 1)
		require.Contains(t, container.Resources.Limits, core.ResourceMemory)
		require.Equal(t, v1, container.Resources.Limits[core.ResourceMemory])
	})

	t.Run("Ensure limits are appended", func(t *testing.T) {
		ApplyContainerResourceRequirements(&container, core.ResourceRequirements{
			Limits: core.ResourceList{
				core.ResourceCPU: v2,
			},
			Requests: core.ResourceList{
				core.ResourceStorage: v2,
			},
		})

		require.Len(t, container.Resources.Requests, 2)
		require.Contains(t, container.Resources.Requests, core.ResourceMemory)
		require.Equal(t, v1, container.Resources.Requests[core.ResourceMemory])

		require.Contains(t, container.Resources.Requests, core.ResourceStorage)
		require.Equal(t, v2, container.Resources.Requests[core.ResourceStorage])

		require.Len(t, container.Resources.Limits, 2)
		require.Contains(t, container.Resources.Limits, core.ResourceMemory)
		require.Equal(t, v1, container.Resources.Limits[core.ResourceMemory])

		require.Contains(t, container.Resources.Limits, core.ResourceCPU)
		require.Equal(t, v2, container.Resources.Limits[core.ResourceCPU])
	})
}
