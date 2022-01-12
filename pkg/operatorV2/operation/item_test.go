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

package operation

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_OperationMatch(t *testing.T) {
	type tc struct {
		Operation Operation

		TestName, Group, Kind, Version, Namespace, Name, Error string
	}

	var cases = []tc{
		{
			TestName: "Empty Item",

			Error: fmt.Sprintf(emptyError, "operation"),
		},
		{
			TestName: "Only operation",

			Operation: Add,

			Error: fmt.Sprintf(emptyError, "version"),
		},
		{
			TestName: "Missing object meta",

			Operation: Add,

			Group:   "test.example",
			Version: "v1alpha",
			Kind:    "test",

			Error: fmt.Sprintf(emptyError, "name"),
		},
		{
			TestName: "With name and namespace",

			Operation: Add,

			Group:   "test.example",
			Version: "v1alpha",
			Kind:    "test",

			Namespace: "default",
			Name:      "test",
		},
		{
			TestName: "With empty namespace",

			Operation: Add,

			Group:   "test.example",
			Version: "v1alpha",
			Kind:    "test",

			Name: "test",
		},
	}

	for _, test := range cases {
		t.Run(test.TestName, func(t *testing.T) {
			i, err := NewItem(test.Operation, test.Group, test.Version, test.Kind, test.Namespace, test.Name)

			if test.Error != "" {
				assert.EqualError(t, err, test.Error)
				return
			}

			require.NoError(t, err)

			result := i.String()

			newI, err := NewItemFromString(result)

			require.NoError(t, err)

			assert.Equal(t, i, newI)

			assert.Equal(t, test.Operation, newI.Operation)
			assert.Equal(t, test.Group, newI.Group)
			assert.Equal(t, test.Version, newI.Version)
			assert.Equal(t, test.Kind, newI.Kind)
			assert.Equal(t, test.Namespace, newI.Namespace)
			assert.Equal(t, test.Name, newI.Name)
		})
	}
}
