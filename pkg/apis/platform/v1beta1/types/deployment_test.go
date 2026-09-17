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

package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_ServiceValues_Hibernated(t *testing.T) {
	t.Run("Hibernated is injected when set", func(t *testing.T) {
		vs, err := Service{
			Platform: ServicePlatform{
				Deployment: ServicePlatformDeployment{Name: "depl"},
				Hibernated: true,
			},
		}.Values()
		require.NoError(t, err)

		m, err := vs.Marshal()
		require.NoError(t, err)

		p, ok := m["arangodb_platform"].(map[string]interface{})
		require.True(t, ok)
		require.Equal(t, true, p["hibernated"])
		require.Equal(t, "depl", p["deployment"].(map[string]interface{})["name"])
	})

	t.Run("Hibernated is omitted by default", func(t *testing.T) {
		vs, err := Service{
			Platform: ServicePlatform{
				Deployment: ServicePlatformDeployment{Name: "depl"},
			},
		}.Values()
		require.NoError(t, err)

		require.NotContains(t, vs.String(), "hibernated")
	})
}
