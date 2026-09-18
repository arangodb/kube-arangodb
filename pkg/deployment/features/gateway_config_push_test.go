//
// DISCLAIMER
//
// Copyright 2026 ArangoDB GmbH, Cologne, Germany
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

package features

import (
	"testing"

	"github.com/stretchr/testify/require"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

func Test_GatewayDynamicModePush(t *testing.T) {
	restore := *gatewayConfigPush.EnabledPointer()
	defer func() { *gatewayConfigPush.EnabledPointer() = restore }()

	dynamicUnset := &api.DeploymentSpecGateway{Enabled: util.NewType(true), Dynamic: util.NewType(true)}
	dynamicPush := &api.DeploymentSpecGateway{Enabled: util.NewType(true), Dynamic: util.NewType(true), DynamicMode: util.NewType(api.GatewayDynamicModePush)}
	dynamicConfigMap := &api.DeploymentSpecGateway{Enabled: util.NewType(true), Dynamic: util.NewType(true), DynamicMode: util.NewType(api.GatewayDynamicModeConfigMap)}
	notDynamic := &api.DeploymentSpecGateway{Enabled: util.NewType(true), Dynamic: util.NewType(false)}

	t.Run("feature enabled", func(t *testing.T) {
		*gatewayConfigPush.EnabledPointer() = true

		require.True(t, GatewayDynamicModePush(dynamicUnset), "push is the default when dynamicMode is unset")
		require.True(t, GatewayDynamicModePush(dynamicPush), "explicit push is honored")
		require.False(t, GatewayDynamicModePush(dynamicConfigMap), "explicit configmap is honored")
		require.False(t, GatewayDynamicModePush(notDynamic), "non-dynamic gateway never pushes")
	})

	t.Run("feature disabled", func(t *testing.T) {
		*gatewayConfigPush.EnabledPointer() = false

		require.False(t, GatewayDynamicModePush(dynamicUnset), "no push when the feature is disabled")
		require.False(t, GatewayDynamicModePush(dynamicPush), "explicit push is forced off when the feature is disabled")
		require.False(t, GatewayDynamicModePush(dynamicConfigMap))
	})
}
