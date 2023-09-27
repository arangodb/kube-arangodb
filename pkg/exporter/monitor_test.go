//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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

package exporter

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_prepareEndpointURL(t *testing.T) {
	tcs := []struct {
		url, path, expected string
	}{
		{"http://some-host", "health", "http://some-host/health"},
		{"https://some-host", "health", "https://some-host/health"},
		{"tcp://some-host", "health", "http://some-host/health"},
		{"ssl://some-host", "health", "https://some-host/health"},
	}

	for i, tc := range tcs {
		u, err := prepareEndpointURL(tc.url, tc.path)
		require.NoErrorf(t, err, "case %d", i)
		require.Equalf(t, tc.expected, u.String(), "case %d", i)
	}
}
