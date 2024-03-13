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

func applyServiceAccount(t *testing.T, template *core.PodTemplateSpec, ns ...*ServiceAccount) func(in func(t *testing.T, pod *core.PodTemplateSpec)) {
	var i *ServiceAccount

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

func Test_ServiceAccount(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		applyServiceAccount(t, nil)(func(t *testing.T, pod *core.PodTemplateSpec) {
			require.Empty(t, pod.Spec.ServiceAccountName)
			require.Nil(t, pod.Spec.AutomountServiceAccountToken)
		})
	})
	t.Run("Empty", func(t *testing.T) {
		applyServiceAccount(t, &core.PodTemplateSpec{})(func(t *testing.T, pod *core.PodTemplateSpec) {
			require.Empty(t, pod.Spec.ServiceAccountName)
			require.Nil(t, pod.Spec.AutomountServiceAccountToken)
		})
	})
	t.Run("One Provided", func(t *testing.T) {
		applyServiceAccount(t, &core.PodTemplateSpec{}, &ServiceAccount{
			ServiceAccountName:           "test",
			AutomountServiceAccountToken: util.NewType(true),
		})(func(t *testing.T, pod *core.PodTemplateSpec) {
			require.EqualValues(t, "test", pod.Spec.ServiceAccountName)

			require.NotNil(t, pod.Spec.AutomountServiceAccountToken)
			require.True(t, *pod.Spec.AutomountServiceAccountToken)
		})
	})
	t.Run("Override", func(t *testing.T) {
		applyServiceAccount(t, &core.PodTemplateSpec{}, &ServiceAccount{
			ServiceAccountName:           "test",
			AutomountServiceAccountToken: util.NewType(true),
		}, &ServiceAccount{
			ServiceAccountName: "test2",
		})(func(t *testing.T, pod *core.PodTemplateSpec) {
			require.EqualValues(t, "test2", pod.Spec.ServiceAccountName)

			require.Nil(t, pod.Spec.AutomountServiceAccountToken)
		})
	})
}
