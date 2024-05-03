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

	"github.com/arangodb/kube-arangodb/pkg/util"
)

func applySecurity(t *testing.T, template *core.PodTemplateSpec, ns ...*Security) func(in func(t *testing.T, pod *core.PodTemplateSpec)) {
	var i *Security

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

func Test_Security(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		applySecurity(t, nil)(func(t *testing.T, pod *core.PodTemplateSpec) {
			require.Nil(t, pod.Spec.SecurityContext)
		})
	})
	t.Run("Empty", func(t *testing.T) {
		applySecurity(t, &core.PodTemplateSpec{})(func(t *testing.T, pod *core.PodTemplateSpec) {
			require.Nil(t, pod.Spec.SecurityContext)
		})
	})
	t.Run("Pick Always Latest Not Nil", func(t *testing.T) {
		applySecurity(t, &core.PodTemplateSpec{}, &Security{
			PodSecurityContext: &core.PodSecurityContext{
				FSGroup:    util.NewType[int64](50),
				RunAsGroup: util.NewType[int64](128),
			},
		}, &Security{
			PodSecurityContext: &core.PodSecurityContext{
				FSGroup: util.NewType[int64](60),
			},
		}, nil)(func(t *testing.T, pod *core.PodTemplateSpec) {
			require.NotNil(t, pod.Spec.SecurityContext)

			require.Nil(t, pod.Spec.SecurityContext.RunAsGroup)
			require.NotNil(t, pod.Spec.SecurityContext.FSGroup)
			require.EqualValues(t, 60, *pod.Spec.SecurityContext.FSGroup)
		})
	})
}
