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

package request_id

import (
	"context"
	"testing"

	pbEnvoyAuthV3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	"github.com/stretchr/testify/require"

	pbImplEnvoyAuthV3Shared "github.com/arangodb/kube-arangodb/integrations/envoy/auth/v3/shared"
	utilConstants "github.com/arangodb/kube-arangodb/pkg/util/constants"
)

func Test_Handle_AddsRequestID(t *testing.T) {
	h, ok, err := New(context.Background(), pbImplEnvoyAuthV3Shared.Configuration{})
	require.NoError(t, err)
	require.True(t, ok)

	current := &pbImplEnvoyAuthV3Shared.Response{}

	require.NoError(t, h.Handle(context.Background(), &pbEnvoyAuthV3.CheckRequest{}, current))

	require.Len(t, current.Headers, 1)
	require.Len(t, current.ResponseHeaders, 1)

	require.Equal(t, utilConstants.EnvoyRequestIDHeader, current.Headers[0].GetHeader().GetKey())
	require.NotEmpty(t, current.Headers[0].GetHeader().GetValue())

	// The request and response carry the same request id.
	require.Equal(t, current.Headers[0].GetHeader().GetValue(), current.ResponseHeaders[0].GetHeader().GetValue())
}
