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

package v2alpha1

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func remarshalVersionWithExpected(t *testing.T, version, expected string) {
	t.Run(version, func(t *testing.T) {
		var v Version

		data, err := json.Marshal(version)
		require.NoError(t, err)

		require.NoError(t, json.Unmarshal(data, &v))

		data, err = json.Marshal(v)
		require.NoError(t, err)

		var newV string

		require.NoError(t, json.Unmarshal(data, &newV))

		require.Equal(t, expected, newV)
	})
}

func Test_Version(t *testing.T) {
	remarshalVersionWithExpected(t, "1", "1.0.0")
	remarshalVersionWithExpected(t, "1.0", "1.0.0")
	remarshalVersionWithExpected(t, "1.0.0", "1.0.0")
	remarshalVersionWithExpected(t, "1.0.0.0", "1.0.0")
	remarshalVersionWithExpected(t, "1.0.0.1", "1.0.0.1")
	remarshalVersionWithExpected(t, "Invalid", "0.0.0")
}
