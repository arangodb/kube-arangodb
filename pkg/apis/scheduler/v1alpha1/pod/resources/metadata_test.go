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
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func applyMetadata(t *testing.T, template *core.PodTemplateSpec, ns ...*Metadata) func(in func(t *testing.T, pod *core.PodTemplateSpec)) {
	var i *Metadata

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

	return func(in func(t *testing.T, spec *core.PodTemplateSpec)) {
		t.Run("Validate", func(t *testing.T) {
			in(t, template)
		})
	}
}

func Test_Metadata(t *testing.T) {
	t.Run("Apply", func(t *testing.T) {
		applyMetadata(t, &core.PodTemplateSpec{})(func(t *testing.T, pod *core.PodTemplateSpec) {
			require.Len(t, pod.Labels, 0)
			require.Len(t, pod.Annotations, 0)
			require.Len(t, pod.OwnerReferences, 0)
		})
	})
	t.Run("Apply nil", func(t *testing.T) {
		applyMetadata(t, nil)(func(t *testing.T, pod *core.PodTemplateSpec) {
			require.Len(t, pod.Labels, 0)
			require.Len(t, pod.Annotations, 0)
			require.Len(t, pod.OwnerReferences, 0)
		})
	})
	t.Run("One", func(t *testing.T) {
		applyMetadata(t, nil, &Metadata{
			Labels: map[string]string{
				"A": "1",
			},
			Annotations: map[string]string{
				"B": "2",
			},
			OwnerReferences: []meta.OwnerReference{
				{
					UID: "test",
				},
			},
		})(func(t *testing.T, pod *core.PodTemplateSpec) {
			require.Len(t, pod.Labels, 1)
			require.Contains(t, pod.Labels, "A")
			require.EqualValues(t, "1", pod.Labels["A"])

			require.Len(t, pod.Annotations, 1)
			require.Contains(t, pod.Annotations, "B")
			require.EqualValues(t, "2", pod.Annotations["B"])

			require.Len(t, pod.OwnerReferences, 1)
			require.EqualValues(t, "test", pod.OwnerReferences[0].UID)
		})
	})
	t.Run("Update", func(t *testing.T) {
		applyMetadata(t, nil, &Metadata{
			Labels: map[string]string{
				"A": "1",
			},
			Annotations: map[string]string{
				"B": "2",
			},
			OwnerReferences: []meta.OwnerReference{
				{
					UID: "test",
				},
			},
		}, &Metadata{
			Labels: map[string]string{
				"A": "3",
			},
			Annotations: map[string]string{
				"B2": "4",
			},
			OwnerReferences: []meta.OwnerReference{
				{
					UID: "test2",
				},
			},
		})(func(t *testing.T, pod *core.PodTemplateSpec) {
			require.Len(t, pod.Labels, 1)
			require.Contains(t, pod.Labels, "A")
			require.EqualValues(t, "3", pod.Labels["A"])

			require.Len(t, pod.Annotations, 2)
			require.Contains(t, pod.Annotations, "B")
			require.EqualValues(t, "2", pod.Annotations["B"])
			require.Contains(t, pod.Annotations, "B2")
			require.EqualValues(t, "4", pod.Annotations["B2"])

			require.Len(t, pod.OwnerReferences, 2)
			require.EqualValues(t, "test", pod.OwnerReferences[0].UID)
			require.EqualValues(t, "test2", pod.OwnerReferences[1].UID)
		})
	})
	t.Run("Update Templat", func(t *testing.T) {
		applyMetadata(t, &core.PodTemplateSpec{
			ObjectMeta: meta.ObjectMeta{
				Labels: map[string]string{
					"A": "1",
				},
				Annotations: map[string]string{
					"B": "2",
				},
				OwnerReferences: []meta.OwnerReference{
					{
						UID: "test",
					},
				},
			},
		}, &Metadata{
			Labels: map[string]string{
				"A": "3",
			},
			Annotations: map[string]string{
				"B2": "4",
			},
			OwnerReferences: []meta.OwnerReference{
				{
					UID: "test2",
				},
			},
		})(func(t *testing.T, pod *core.PodTemplateSpec) {
			require.Len(t, pod.Labels, 1)
			require.Contains(t, pod.Labels, "A")
			require.EqualValues(t, "3", pod.Labels["A"])

			require.Len(t, pod.Annotations, 2)
			require.Contains(t, pod.Annotations, "B")
			require.EqualValues(t, "2", pod.Annotations["B"])
			require.Contains(t, pod.Annotations, "B2")
			require.EqualValues(t, "4", pod.Annotations["B2"])

			require.Len(t, pod.OwnerReferences, 2)
			require.EqualValues(t, "test", pod.OwnerReferences[0].UID)
			require.EqualValues(t, "test2", pod.OwnerReferences[1].UID)
		})
	})
}
