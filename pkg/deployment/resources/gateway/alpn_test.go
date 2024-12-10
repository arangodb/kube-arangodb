//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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

package gateway

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_ALPN(t *testing.T) {
	require.Equal(t, "", ALPNProtocol(0).String())
	require.Equal(t, "http/1.1", ALPNProtocolHTTP1.String())
	require.Equal(t, "h2", ALPNProtocolHTTP2.String())
	require.Equal(t, "h2,http/1.1", (ALPNProtocolHTTP1 | ALPNProtocolHTTP2).String())
	require.Equal(t, "h2,http/1.1", (ALPNProtocolHTTP2 | ALPNProtocolHTTP1).String())
}
