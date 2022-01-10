//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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

package v1

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type checksumCompareCases []checksumCompareCase

type checksumCompareCase struct {
	name string

	spec     DeploymentSpec
	checksum string
}

func runChecksumCompareCases(t *testing.T, cases checksumCompareCases) {
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			runChecksumCompareCase(t, c)
		})
	}
}

func runChecksumCompareCase(t *testing.T, c checksumCompareCase) {
	s, err := c.spec.Checksum()
	require.NoError(t, err)

	require.Equalf(t, c.checksum, s, "Checksum od ArangoDeployment mismatch")
}

func TestImmutableSpec(t *testing.T) {
	cases := checksumCompareCases{
		{
			name:     "Default case - from 1.0.3",
			spec:     DeploymentSpec{},
			checksum: "a164088b280d72c177c2eafdab7a346fb296264b70c06329b776c506925bb54e",
		},
	}

	runChecksumCompareCases(t, cases)
}
