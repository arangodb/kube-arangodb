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

package auth_required

import (
	"context"
	goHttp "net/http"
	"testing"

	pbEnvoyAuthV3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	"github.com/stretchr/testify/require"

	pbImplEnvoyAuthV3Shared "github.com/arangodb/kube-arangodb/integrations/envoy/auth/v3/shared"
)

func requestWithExtensions(ext map[string]string) *pbEnvoyAuthV3.CheckRequest {
	return &pbEnvoyAuthV3.CheckRequest{
		Attributes: &pbEnvoyAuthV3.AttributeContext{
			ContextExtensions: ext,
		},
	}
}

func Test_New_Disabled(t *testing.T) {
	_, ok, err := New(context.Background(), pbImplEnvoyAuthV3Shared.Configuration{Enabled: false})
	require.NoError(t, err)
	require.False(t, ok)
}

func Test_Handle(t *testing.T) {
	h, ok, err := New(context.Background(), pbImplEnvoyAuthV3Shared.Configuration{Enabled: true})
	require.NoError(t, err)
	require.True(t, ok)

	authenticated := &pbImplEnvoyAuthV3Shared.Response{User: &pbImplEnvoyAuthV3Shared.ResponseAuth{User: "root"}}
	anonymous := &pbImplEnvoyAuthV3Shared.Response{}

	requiredExt := map[string]string{
		pbImplEnvoyAuthV3Shared.AuthConfigAuthRequiredKey: pbImplEnvoyAuthV3Shared.AuthConfigKeywordTrue,
	}

	t.Run("Not required, anonymous is allowed", func(t *testing.T) {
		require.NoError(t, h.Handle(context.Background(), requestWithExtensions(nil), anonymous))
	})

	t.Run("Required, anonymous is denied", func(t *testing.T) {
		err := h.Handle(context.Background(), requestWithExtensions(requiredExt), anonymous)
		var denied pbImplEnvoyAuthV3Shared.DeniedResponse
		require.ErrorAs(t, err, &denied)
		require.EqualValues(t, goHttp.StatusUnauthorized, denied.Code)
	})

	t.Run("Required, authenticated is allowed", func(t *testing.T) {
		require.NoError(t, h.Handle(context.Background(), requestWithExtensions(requiredExt), authenticated))
	})
}
