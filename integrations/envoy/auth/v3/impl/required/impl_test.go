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

package required

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

func Test_Handle(t *testing.T) {
	h, ok, err := New(context.Background(), pbImplEnvoyAuthV3Shared.Configuration{})
	require.NoError(t, err)
	require.True(t, ok)

	t.Run("Missing type key is denied", func(t *testing.T) {
		err := h.Handle(context.Background(), requestWithExtensions(nil), &pbImplEnvoyAuthV3Shared.Response{})
		var denied pbImplEnvoyAuthV3Shared.DeniedResponse
		require.ErrorAs(t, err, &denied)
		require.EqualValues(t, goHttp.StatusBadRequest, denied.Code)
	})

	t.Run("Wrong type value is denied", func(t *testing.T) {
		err := h.Handle(context.Background(), requestWithExtensions(map[string]string{
			pbImplEnvoyAuthV3Shared.AuthConfigTypeKey: "other",
		}), &pbImplEnvoyAuthV3Shared.Response{})
		var denied pbImplEnvoyAuthV3Shared.DeniedResponse
		require.ErrorAs(t, err, &denied)
	})

	t.Run("Matching type is allowed", func(t *testing.T) {
		err := h.Handle(context.Background(), requestWithExtensions(map[string]string{
			pbImplEnvoyAuthV3Shared.AuthConfigTypeKey: pbImplEnvoyAuthV3Shared.AuthConfigTypeValue,
		}), &pbImplEnvoyAuthV3Shared.Response{})
		require.NoError(t, err)
	})
}
