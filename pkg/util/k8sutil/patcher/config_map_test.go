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

package patcher

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/util/tests"
)

func Test_ConfigMap(t *testing.T) {
	c := tests.NewEmptyInspector(t)

	t.Run("Create", func(t *testing.T) {
		require.NoError(t, c.Refresh(context.Background()))

		_, err := c.ConfigMapsModInterface().V1().Create(context.Background(), &core.ConfigMap{
			ObjectMeta: meta.ObjectMeta{
				Name:      "test",
				Namespace: c.Namespace(),
			},
		}, meta.CreateOptions{})
		require.NoError(t, err)
	})

	require.NoError(t, c.Refresh(context.Background()))

	t.Run("Check", func(t *testing.T) {
		cm, ok := c.ConfigMap().V1().GetSimple("test")
		require.True(t, ok)
		require.Len(t, cm.Data, 0)
	})

	require.NoError(t, c.Refresh(context.Background()))

	t.Run("Update", func(t *testing.T) {
		cm, ok := c.ConfigMap().V1().GetSimple("test")
		require.True(t, ok)
		uCm, ok, err := Patcher[*core.ConfigMap](context.Background(), c.ConfigMapsModInterface().V1(), cm, meta.PatchOptions{}, PatchConfigMapData(map[string]string{
			"A": "B",
		}))
		require.NoError(t, err)
		require.True(t, ok)

		require.NoError(t, c.Refresh(context.Background()))

		cm, ok = c.ConfigMap().V1().GetSimple("test")
		require.True(t, ok)

		require.Equal(t, map[string]string{
			"A": "B",
		}, uCm.Data)

		require.Equal(t, map[string]string{
			"A": "B",
		}, cm.Data)
	})

	t.Run("Reupdate", func(t *testing.T) {
		cm, ok := c.ConfigMap().V1().GetSimple("test")
		require.True(t, ok)

		uCm, ok, err := Patcher[*core.ConfigMap](context.Background(), c.ConfigMapsModInterface().V1(), cm, meta.PatchOptions{}, PatchConfigMapData(map[string]string{
			"A": "B",
		}))
		require.NoError(t, err)
		require.False(t, ok)

		require.NoError(t, c.Refresh(context.Background()))

		cm, ok = c.ConfigMap().V1().GetSimple("test")
		require.True(t, ok)

		require.Equal(t, map[string]string{
			"A": "B",
		}, uCm.Data)

		require.Equal(t, map[string]string{
			"A": "B",
		}, cm.Data)
	})
}
