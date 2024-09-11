//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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

	schedulerPolicyApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1/policy"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

func applyCore(t *testing.T, template *core.PodTemplateSpec, container *core.Container, ns ...*Core) func(in func(t *testing.T, pod *core.PodTemplateSpec, container *core.Container)) {
	var i *Core

	for _, n := range ns {
		require.NoError(t, n.Validate())

		i = i.With(n)

		require.NoError(t, i.Validate())
	}

	template = template.DeepCopy()

	if template == nil {
		template = &core.PodTemplateSpec{}
	}

	container = container.DeepCopy()
	if container == nil {
		container = &core.Container{}
	}

	template.Spec.Containers = append(template.Spec.Containers, *container)

	container = &template.Spec.Containers[0]

	require.NoError(t, i.Apply(template, container))

	return func(in func(t *testing.T, spec *core.PodTemplateSpec, container *core.Container)) {
		t.Run("Validate", func(t *testing.T) {
			in(t, template, container)
		})
	}
}

func Test_Core(t *testing.T) {
	t.Run("With Nil", func(t *testing.T) {
		applyCore(t, nil, nil)(func(t *testing.T, _ *core.PodTemplateSpec, container *core.Container) {
			require.Len(t, container.Args, 0)
			require.Len(t, container.Command, 0)
		})
	})
	t.Run("With Empty", func(t *testing.T) {
		applyCore(t, &core.PodTemplateSpec{}, &core.Container{})(func(t *testing.T, _ *core.PodTemplateSpec, container *core.Container) {
			require.Len(t, container.Args, 0)
			require.Len(t, container.Command, 0)
		})
	})
	t.Run("With One", func(t *testing.T) {
		applyCore(t, &core.PodTemplateSpec{}, &core.Container{}, &Core{
			Command:    []string{"A"},
			Args:       []string{"B"},
			WorkingDir: "C",
		})(func(t *testing.T, _ *core.PodTemplateSpec, container *core.Container) {
			require.Len(t, container.Args, 1)
			require.Contains(t, container.Args, "B")

			require.Len(t, container.Command, 1)
			require.Contains(t, container.Command, "A")

			require.EqualValues(t, "C", container.WorkingDir)
		})
	})
	t.Run("With Override", func(t *testing.T) {
		applyCore(t, &core.PodTemplateSpec{}, &core.Container{}, &Core{
			Command: []string{"B"},
			Args:    []string{"A"},
		})(func(t *testing.T, _ *core.PodTemplateSpec, container *core.Container) {
			require.Len(t, container.Args, 1)
			require.Contains(t, container.Args, "A")

			require.Len(t, container.Command, 1)
			require.Contains(t, container.Command, "B")

			require.EqualValues(t, "", container.WorkingDir)
		})
	})
	t.Run("With Append", func(t *testing.T) {
		applyCore(t, &core.PodTemplateSpec{}, &core.Container{}, &Core{
			Command: []string{"B"},
			Args:    []string{"A"},
		}, &Core{
			Policy: &schedulerPolicyApi.Policy{
				Method: util.NewType(schedulerPolicyApi.Append),
			},
			Command: []string{"C"},
			Args:    []string{"D"},
		})(func(t *testing.T, _ *core.PodTemplateSpec, container *core.Container) {
			require.Len(t, container.Args, 2)
			require.Contains(t, container.Args, "A")
			require.Contains(t, container.Args, "D")
			require.Equal(t, []string{"A", "D"}, container.Args)

			require.Len(t, container.Command, 1)
			require.Contains(t, container.Command, "C")

			require.EqualValues(t, "", container.WorkingDir)
		})
	})
}
