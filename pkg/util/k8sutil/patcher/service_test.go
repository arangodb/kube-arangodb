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

package patcher

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/util/tests"
)

func Test_Service(t *testing.T) {
	c := tests.NewEmptyInspector(t)

	t.Run("Create", func(t *testing.T) {
		require.NoError(t, c.Refresh(context.Background()))

		_, err := c.ServicesModInterface().V1().Create(context.Background(), &core.Service{
			ObjectMeta: meta.ObjectMeta{
				Name:      "test",
				Namespace: c.Namespace(),
			},
		}, meta.CreateOptions{})
		require.NoError(t, err)
	})

	t.Run("publishNotReadyAddresses", func(t *testing.T) {
		t.Run("Set to true", func(t *testing.T) {
			require.NoError(t, c.Refresh(context.Background()))
			svc, ok := c.Service().V1().GetSimple("test")
			require.True(t, ok)
			require.False(t, svc.Spec.PublishNotReadyAddresses)

			changed, err := ServicePatcher(context.Background(), c.ServicesModInterface().V1(), svc, meta.PatchOptions{}, PatchServicePublishNotReadyAddresses(true))
			require.NoError(t, err)
			require.True(t, changed)

			require.NoError(t, c.Refresh(context.Background()))
			svc, ok = c.Service().V1().GetSimple("test")
			require.True(t, ok)
			require.True(t, svc.Spec.PublishNotReadyAddresses)
		})

		t.Run("Reset to true", func(t *testing.T) {
			require.NoError(t, c.Refresh(context.Background()))
			svc, ok := c.Service().V1().GetSimple("test")
			require.True(t, ok)
			require.True(t, svc.Spec.PublishNotReadyAddresses)

			changed, err := ServicePatcher(context.Background(), c.ServicesModInterface().V1(), svc, meta.PatchOptions{}, PatchServicePublishNotReadyAddresses(true))
			require.NoError(t, err)
			require.False(t, changed)

			require.NoError(t, c.Refresh(context.Background()))
			svc, ok = c.Service().V1().GetSimple("test")
			require.True(t, ok)
			require.True(t, svc.Spec.PublishNotReadyAddresses)
		})

		t.Run("Set to false", func(t *testing.T) {
			require.NoError(t, c.Refresh(context.Background()))
			svc, ok := c.Service().V1().GetSimple("test")
			require.True(t, ok)
			require.True(t, svc.Spec.PublishNotReadyAddresses)

			changed, err := ServicePatcher(context.Background(), c.ServicesModInterface().V1(), svc, meta.PatchOptions{}, PatchServicePublishNotReadyAddresses(false))
			require.NoError(t, err)
			require.True(t, changed)

			require.NoError(t, c.Refresh(context.Background()))
			svc, ok = c.Service().V1().GetSimple("test")
			require.True(t, ok)
			require.False(t, svc.Spec.PublishNotReadyAddresses)
		})
	})
}
