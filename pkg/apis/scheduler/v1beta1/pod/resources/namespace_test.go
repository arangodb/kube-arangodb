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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	core "k8s.io/api/core/v1"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

func applyNamespace(t *testing.T, template *core.PodTemplateSpec, ns ...*Namespace) func(in func(t *testing.T, pod *core.PodTemplateSpec)) {
	var i *Namespace

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

func Test_Namespace(t *testing.T) {
	t.Run("Apply", func(t *testing.T) {
		applyNamespace(t, &core.PodTemplateSpec{})(func(t *testing.T, pod *core.PodTemplateSpec) {
			assert.False(t, pod.Spec.HostNetwork)
			assert.False(t, pod.Spec.HostPID)
			assert.False(t, pod.Spec.HostIPC)
			assert.Nil(t, pod.Spec.ShareProcessNamespace)
		})
	})
	t.Run("Apply nil", func(t *testing.T) {
		applyNamespace(t, nil)(func(t *testing.T, pod *core.PodTemplateSpec) {
			assert.False(t, pod.Spec.HostNetwork)
			assert.False(t, pod.Spec.HostPID)
			assert.False(t, pod.Spec.HostIPC)
			assert.Nil(t, pod.Spec.ShareProcessNamespace)
		})
	})
	t.Run("Apply with nils", func(t *testing.T) {
		applyNamespace(t, nil)(func(t *testing.T, pod *core.PodTemplateSpec) {
			assert.False(t, pod.Spec.HostNetwork)
			assert.False(t, pod.Spec.HostPID)
			assert.False(t, pod.Spec.HostIPC)
			assert.Nil(t, pod.Spec.ShareProcessNamespace)
		})
	})
	t.Run("Apply with template", func(t *testing.T) {
		applyNamespace(t, nil, &Namespace{
			HostPID: util.NewType(true),
		})(func(t *testing.T, pod *core.PodTemplateSpec) {
			assert.False(t, pod.Spec.HostNetwork)
			assert.True(t, pod.Spec.HostPID)
			assert.False(t, pod.Spec.HostIPC)
			assert.Nil(t, pod.Spec.ShareProcessNamespace)
		})
	})
	t.Run("Apply with template overrides", func(t *testing.T) {
		applyNamespace(t, nil, &Namespace{
			HostPID: util.NewType(true),
		}, &Namespace{
			HostNetwork: util.NewType(true),
		}, &Namespace{
			HostPID: util.NewType(false),
		})(func(t *testing.T, pod *core.PodTemplateSpec) {
			assert.True(t, pod.Spec.HostNetwork)
			assert.False(t, pod.Spec.HostPID)
			assert.False(t, pod.Spec.HostIPC)
			assert.Nil(t, pod.Spec.ShareProcessNamespace)
		})
	})
	t.Run("Apply with all", func(t *testing.T) {
		applyNamespace(t, nil, &Namespace{
			HostNetwork:           util.NewType(true),
			HostPID:               util.NewType(true),
			HostIPC:               util.NewType(true),
			ShareProcessNamespace: util.NewType(true),
		})(func(t *testing.T, pod *core.PodTemplateSpec) {
			assert.True(t, pod.Spec.HostNetwork)
			assert.True(t, pod.Spec.HostPID)
			assert.True(t, pod.Spec.HostIPC)
			assert.NotNil(t, pod.Spec.ShareProcessNamespace)
			assert.True(t, *pod.Spec.ShareProcessNamespace)
		})
	})
	t.Run("Apply false", func(t *testing.T) {
		applyNamespace(t, nil, &Namespace{
			HostNetwork:           util.NewType(true),
			HostPID:               util.NewType(true),
			HostIPC:               util.NewType(true),
			ShareProcessNamespace: util.NewType(false),
		})(func(t *testing.T, pod *core.PodTemplateSpec) {
			assert.True(t, pod.Spec.HostNetwork)
			assert.True(t, pod.Spec.HostPID)
			assert.True(t, pod.Spec.HostIPC)
			assert.NotNil(t, pod.Spec.ShareProcessNamespace)
			assert.False(t, *pod.Spec.ShareProcessNamespace)
		})
	})
}
