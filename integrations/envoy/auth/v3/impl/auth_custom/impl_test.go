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

package auth_custom

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	pbImplEnvoyAuthV3Shared "github.com/arangodb/kube-arangodb/integrations/envoy/auth/v3/shared"
)

func Test_New(t *testing.T) {
	ctx := context.Background()

	t.Run("Disabled", func(t *testing.T) {
		_, ok, err := New(ctx, pbImplEnvoyAuthV3Shared.Configuration{Enabled: false})
		require.NoError(t, err)
		require.False(t, ok)
	})

	t.Run("Auth disabled", func(t *testing.T) {
		_, ok, err := New(ctx, pbImplEnvoyAuthV3Shared.Configuration{
			Enabled: true,
			Auth:    pbImplEnvoyAuthV3Shared.ConfigurationAuth{Enabled: false},
		})
		require.NoError(t, err)
		require.False(t, ok)
	})

	t.Run("Unknown type", func(t *testing.T) {
		_, ok, err := New(ctx, pbImplEnvoyAuthV3Shared.Configuration{
			Enabled: true,
			Auth:    pbImplEnvoyAuthV3Shared.ConfigurationAuth{Enabled: true, Type: "Unknown"},
		})
		require.NoError(t, err)
		require.False(t, ok)
	})

	t.Run("OpenID requires a database client in context", func(t *testing.T) {
		_, ok, err := New(ctx, pbImplEnvoyAuthV3Shared.Configuration{
			Enabled: true,
			Auth:    pbImplEnvoyAuthV3Shared.ConfigurationAuth{Enabled: true, Type: "OpenID"},
		})
		require.Error(t, err)
		require.False(t, ok)
	})
}
