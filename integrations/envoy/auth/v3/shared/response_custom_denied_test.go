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

package shared

import (
	goHttp "net/http"
	"testing"

	pbEnvoyCoreV3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	"github.com/stretchr/testify/require"
)

func headerValue(headers []*pbEnvoyCoreV3.HeaderValueOption, key string) (string, bool) {
	for _, h := range headers {
		if h.GetHeader().GetKey() == key {
			return h.GetHeader().GetValue(), true
		}
	}
	return "", false
}

func Test_DeniedResponse_Error(t *testing.T) {
	require.Equal(t, "Request denied with code: 401", DeniedResponse{Code: goHttp.StatusUnauthorized}.Error())
}

func Test_DeniedResponse_Response(t *testing.T) {
	t.Run("Status code and no body", func(t *testing.T) {
		resp, err := DeniedResponse{Code: goHttp.StatusForbidden}.Response()
		require.NoError(t, err)

		require.EqualValues(t, goHttp.StatusForbidden, resp.GetStatus().GetCode())

		denied := resp.GetDeniedResponse()
		require.NotNil(t, denied)
		require.EqualValues(t, goHttp.StatusForbidden, denied.GetStatus().GetCode())
		require.Empty(t, denied.GetBody())
	})

	t.Run("Message is rendered as JSON body with content type", func(t *testing.T) {
		resp, err := DeniedResponse{
			Code:    goHttp.StatusUnauthorized,
			Message: &DeniedMessage{Message: "Unauthorized"},
		}.Response()
		require.NoError(t, err)

		denied := resp.GetDeniedResponse()
		require.NotNil(t, denied)
		require.JSONEq(t, `{"message":"Unauthorized"}`, denied.GetBody())

		ct, ok := headerValue(denied.GetHeaders(), "content-type")
		require.True(t, ok)
		require.Equal(t, "application/json", ct)
	})

	t.Run("Custom headers are propagated", func(t *testing.T) {
		resp, err := DeniedResponse{
			Code:    goHttp.StatusUnauthorized,
			Headers: map[string]string{"WWW-Authenticate": "Bearer"},
		}.Response()
		require.NoError(t, err)

		v, ok := headerValue(resp.GetDeniedResponse().GetHeaders(), "WWW-Authenticate")
		require.True(t, ok)
		require.Equal(t, "Bearer", v)
	})
}
