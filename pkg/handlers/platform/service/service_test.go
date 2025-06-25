//
// DISCLAIMER
//
// Copyright 2025 ArangoDB GmbH, Cologne, Germany
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

package service

import (
	"testing"

	"github.com/stretchr/testify/require"

	platformApi "github.com/arangodb/kube-arangodb/pkg/apis/platform/v1alpha1"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"
	"github.com/arangodb/kube-arangodb/pkg/util/tests"
)

func Test_ServiceReconcilee(t *testing.T) {
	handler := newFakeHandler()

	// Arrange
	extension := tests.NewMetaObject[*platformApi.ArangoPlatformService](t, tests.FakeNamespace, "example",
		func(t *testing.T, obj *platformApi.ArangoPlatformService) {})

	refresh := tests.CreateObjects(t, handler.kubeClient, handler.client, &extension)

	t.Run("Missing chart", func(t *testing.T) {
		// Test
		require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

		// Refresh
		refresh(t)

		// Validate
		require.False(t, extension.Status.Conditions.IsTrue(platformApi.SpecValidCondition))
		require.False(t, extension.Status.Conditions.IsTrue(platformApi.ReadyCondition))
	})
}
