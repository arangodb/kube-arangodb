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

package pod

import (
	"testing"

	"github.com/stretchr/testify/require"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/yaml"

	schedulerPodResourcesApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1alpha1/pod/resources"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

func applyPod(t *testing.T, template *core.PodTemplateSpec, ns ...*Pod) func(in func(t *testing.T, pod *core.PodTemplateSpec, podData *Pod)) {
	var i *Pod

	for _, n := range ns {
		require.NoError(t, n.Validate())

		i = i.With(n)

		require.NoError(t, i.Validate())
	}

	template = template.DeepCopy()

	if template == nil {
		template = &core.PodTemplateSpec{}
	}

	require.NoError(t, i.Apply(template))

	return func(in func(t *testing.T, spec *core.PodTemplateSpec, podData *Pod)) {
		t.Run("Validate", func(t *testing.T) {
			if i != nil {
				in(t, template, i)
			} else {
				in(t, template, &Pod{})
			}
		})
	}
}

func applyPodYAML(t *testing.T, template *core.PodTemplateSpec, ns ...string) func(in func(t *testing.T, pod *core.PodTemplateSpec, podData *Pod)) {
	elements := make([]*Pod, len(ns))

	for id := range ns {
		var p Pod
		require.NoError(t, yaml.Unmarshal([]byte(ns[id]), &p))
		elements[id] = p.DeepCopy()
	}

	return applyPod(t, template, elements...)
}

func Test_Pod(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		applyPod(t, nil)(func(t *testing.T, pod *core.PodTemplateSpec, _ *Pod) {
			require.Nil(t, pod.Spec.SecurityContext)
			require.Nil(t, pod.Spec.Affinity)
			require.Empty(t, pod.Spec.ServiceAccountName)
		})
	})
	t.Run("Empty template", func(t *testing.T) {
		applyPod(t, &core.PodTemplateSpec{})(func(t *testing.T, pod *core.PodTemplateSpec, _ *Pod) {
			require.Nil(t, pod.Spec.SecurityContext)
			require.Nil(t, pod.Spec.Affinity)
			require.Empty(t, pod.Spec.ServiceAccountName)
		})
	})
	t.Run("Add scheduling", func(t *testing.T) {
		applyPod(t, &core.PodTemplateSpec{}, &Pod{
			Security: &schedulerPodResourcesApi.Security{
				PodSecurityContext: &core.PodSecurityContext{
					RunAsGroup: util.NewType[int64](50),
				},
			},
		}, &Pod{
			Scheduling: &schedulerPodResourcesApi.Scheduling{
				NodeSelector: map[string]string{
					"A": "B",
				},
			},
		}, &Pod{
			Scheduling: &schedulerPodResourcesApi.Scheduling{
				NodeSelector: map[string]string{
					"A1": "B1",
				},
			},
		})(func(t *testing.T, pod *core.PodTemplateSpec, _ *Pod) {
			require.NotNil(t, pod.Spec.SecurityContext)
			require.NotNil(t, pod.Spec.SecurityContext.RunAsGroup)
			require.EqualValues(t, 50, *pod.Spec.SecurityContext.RunAsGroup)

			require.NotNil(t, pod.Spec.NodeSelector)
			require.Len(t, pod.Spec.NodeSelector, 2)

			require.Contains(t, pod.Spec.NodeSelector, "A")
			require.Equal(t, "B", pod.Spec.NodeSelector["A"])

			require.Contains(t, pod.Spec.NodeSelector, "A1")
			require.Equal(t, "B1", pod.Spec.NodeSelector["A1"])
		})
	})
}

func Test_Pod_YAML(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		applyPodYAML(t, nil)(func(t *testing.T, _ *core.PodTemplateSpec, spec *Pod) {
			require.Nil(t, spec.Security)
			require.Nil(t, spec.Scheduling)
			require.Nil(t, spec.Namespace)
			require.Nil(t, spec.ServiceAccount)
			require.Nil(t, spec.Metadata)
		})
	})
	t.Run("Empty template", func(t *testing.T) {
		applyPodYAML(t, &core.PodTemplateSpec{})(func(t *testing.T, pod *core.PodTemplateSpec, _ *Pod) {
			require.Nil(t, pod.Spec.SecurityContext)
			require.Nil(t, pod.Spec.Affinity)
			require.Empty(t, pod.Spec.ServiceAccountName)
			require.Nil(t, pod.Labels)
		})
	})
	t.Run("Add nodeSelector", func(t *testing.T) {
		applyPodYAML(t, nil, `
---

nodeSelector:
  A: B
`)(func(t *testing.T, _ *core.PodTemplateSpec, spec *Pod) {
			require.Nil(t, spec.Security)
			require.NotNil(t, spec.Scheduling)
			require.Len(t, spec.Scheduling.NodeSelector, 1)
			require.Contains(t, spec.Scheduling.NodeSelector, "A")
			require.Equal(t, "B", spec.Scheduling.NodeSelector["A"])
			require.Nil(t, spec.Namespace)
		})
	})
	t.Run("Merge nodeSelector", func(t *testing.T) {
		applyPodYAML(t, nil, `
---

nodeSelector:
  A: B
`, `
---

nodeSelector:
  C: D
`)(func(t *testing.T, _ *core.PodTemplateSpec, spec *Pod) {
			require.Nil(t, spec.Security)
			require.NotNil(t, spec.Scheduling)
			require.Len(t, spec.Scheduling.NodeSelector, 2)
			require.Contains(t, spec.Scheduling.NodeSelector, "A")
			require.Equal(t, "B", spec.Scheduling.NodeSelector["A"])
			require.Contains(t, spec.Scheduling.NodeSelector, "C")
			require.Equal(t, "D", spec.Scheduling.NodeSelector["C"])
			require.Nil(t, spec.Namespace)
		})
	})
	t.Run("Add multiple values", func(t *testing.T) {
		applyPodYAML(t, nil, `
---

nodeSelector:
  A: B
podSecurityContext:
  runAsUser: 10
hostPID: true
serviceAccountName: test
labels:
  A: B
volumes:
  - name: test
    emptyDir: {}
`)(func(t *testing.T, pod *core.PodTemplateSpec, spec *Pod) {
			// Spec
			require.NotNil(t, spec.Security)
			require.NotNil(t, spec.Security.PodSecurityContext)
			require.NotNil(t, spec.Security.PodSecurityContext.RunAsUser)
			require.EqualValues(t, 10, *spec.Security.PodSecurityContext.RunAsUser)
			require.NotNil(t, spec.Scheduling)
			require.Len(t, spec.Scheduling.NodeSelector, 1)
			require.Contains(t, spec.Scheduling.NodeSelector, "A")
			require.Equal(t, "B", spec.Scheduling.NodeSelector["A"])
			require.NotNil(t, spec.Namespace)
			require.NotNil(t, spec.Namespace.HostPID)
			require.True(t, *spec.Namespace.HostPID)
			require.NotNil(t, spec.Volumes)
			require.Len(t, spec.Volumes.Volumes, 1)
			require.EqualValues(t, "test", spec.Volumes.Volumes[0].Name)
			require.NotNil(t, spec.Volumes.Volumes[0].EmptyDir)
			require.NotNil(t, spec.ServiceAccount)
			require.EqualValues(t, "test", spec.ServiceAccount.ServiceAccountName)
			require.NotNil(t, spec.Labels)
			require.Contains(t, spec.Labels, "A")
			require.EqualValues(t, "B", spec.Labels["A"])

			// Pod
			require.NotNil(t, pod.Spec.SecurityContext)
			require.NotNil(t, pod.Spec.SecurityContext.RunAsUser)
			require.EqualValues(t, 10, *pod.Spec.SecurityContext.RunAsUser)
			require.NotNil(t, pod.Spec.NodeSelector)
			require.Len(t, pod.Spec.NodeSelector, 1)
			require.Contains(t, pod.Spec.NodeSelector, "A")
			require.Equal(t, "B", pod.Spec.NodeSelector["A"])
			require.Nil(t, pod.Spec.Affinity)
			require.NotNil(t, pod.Spec.HostPID)
			require.True(t, pod.Spec.HostPID)
			require.Len(t, pod.Spec.Volumes, 1)
			require.EqualValues(t, "test", pod.Spec.Volumes[0].Name)
			require.NotNil(t, pod.Spec.Volumes[0].EmptyDir)
			require.NotNil(t, pod.Labels)
			require.Contains(t, pod.Labels, "A")
			require.EqualValues(t, "B", spec.Labels["A"])
		})
	})
}
