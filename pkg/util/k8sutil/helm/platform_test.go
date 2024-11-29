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

package helm

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/arangodb/kube-arangodb/pkg/util/tests/suite"
)

func Test_Platform(t *testing.T) {
	c1, err := Chart(suite.GetChart(t, "example", "1.0.0")).Get()
	require.NoError(t, err)

	p1, err := c1.Platform()
	require.NoError(t, err)

	c2, err := Chart(suite.GetChart(t, "example", "1.0.1")).Get()
	require.NoError(t, err)

	p2, err := c2.Platform()
	require.NoError(t, err)

	require.Nil(t, p1)
	require.NotNil(t, p2)
}
