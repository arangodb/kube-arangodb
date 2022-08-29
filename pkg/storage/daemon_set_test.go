//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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

package storage

import (
	"testing"

	"github.com/stretchr/testify/require"
	core "k8s.io/api/core/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/storage/v1alpha"
)

// TestEnsureDaemonSet tests ensureDaemonSet() method
func TestEnsureDaemonSet(t *testing.T) {
	testImage := "test-image"

	tps := []core.LocalObjectReference{
		{
			Name: "custom-docker",
		},
	}

	ls, ds := generateDaemonSet(t, core.PodSpec{
		ImagePullSecrets: tps,
		Containers: []core.Container{
			{
				Name:            testImage,
				ImagePullPolicy: core.PullAlways,
				Image:           testImage,
			},
		},
	}, api.LocalStorageSpec{})

	require.Equal(t, ds.GetName(), ls.apiObject.GetName())
	require.Equal(t, ds.Spec.Template.Spec.ImagePullSecrets, tps)
	require.Equal(t, len(ds.Spec.Template.Spec.Containers), 1)

	c := ds.Spec.Template.Spec.Containers[0]
	require.Equal(t, c.Image, testImage)
	require.Equal(t, c.ImagePullPolicy, core.PullAlways)
	require.Nil(t, ds.Spec.Template.Spec.Priority)
}

// TestEnsureDaemonSet tests ensureDaemonSet() method
func TestEnsureDaemonSet_WithPriority(t *testing.T) {
	testImage := "test-image"
	var priority int32 = 555

	tps := []core.LocalObjectReference{
		{
			Name: "custom-docker",
		},
	}

	ls, ds := generateDaemonSet(t, core.PodSpec{
		ImagePullSecrets: tps,
		Containers: []core.Container{
			{
				Name:            testImage,
				ImagePullPolicy: core.PullAlways,
				Image:           testImage,
			},
		},
	}, api.LocalStorageSpec{
		PodCustomization: &api.LocalStoragePodCustomization{
			Priority: &priority,
		},
	})

	require.Equal(t, ds.GetName(), ls.apiObject.GetName())
	require.Equal(t, ds.Spec.Template.Spec.ImagePullSecrets, tps)
	require.Equal(t, len(ds.Spec.Template.Spec.Containers), 1)

	c := ds.Spec.Template.Spec.Containers[0]
	require.Equal(t, c.Image, testImage)
	require.Equal(t, c.ImagePullPolicy, core.PullAlways)
	require.NotNil(t, ds.Spec.Template.Spec.Priority)
	require.Equal(t, priority, *ds.Spec.Template.Spec.Priority)
}
