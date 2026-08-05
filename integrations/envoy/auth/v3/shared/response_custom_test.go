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
	"testing"

	pbEnvoyAuthV3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	"github.com/stretchr/testify/require"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func Test_CustomStaticResponse(t *testing.T) {
	checkResponse := &pbEnvoyAuthV3.CheckResponse{}

	err := NewCustomStaticResponse(checkResponse)
	require.Error(t, err)
	require.Equal(t, "Request static response", err.Error())

	// The error must be recoverable as a CustomResponse, as the service dispatch relies on it.
	var custom CustomResponse
	require.True(t, errors.As(err, &custom))

	got, gErr := custom.Response()
	require.NoError(t, gErr)
	require.Same(t, checkResponse, got)
}
