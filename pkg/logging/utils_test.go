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

package logging

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_ParseLogLevelsFromArgs(t *testing.T) {
	type testCase struct {
		name string
		in   []string

		expectedErr string
		expected    map[string]Level
	}

	testCases := []testCase{
		{
			name: "empty",

			expected: map[string]Level{},
		},
		{
			name: "default level",

			in: []string{
				"info",
			},

			expected: map[string]Level{
				TopicAll: Info,
			},
		},
		{
			name: "parse error level",

			in: []string{
				"infxx",
			},

			expectedErr: "Unknown Level String: 'infxx', defaulting to NoLevel",
		},
		{
			name: "default level - camel",

			in: []string{
				"iNfO",
			},

			expected: map[string]Level{
				TopicAll: Info,
			},
		},
		{
			name: "default level + specific",

			in: []string{
				"iNfO",
				"other=debug",
			},

			expected: map[string]Level{
				TopicAll: Info,
				"other":  Debug,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r, err := ParseLogLevelsFromArgs(tc.in)

			if tc.expectedErr != "" {
				require.EqualError(t, err, tc.expectedErr)
				require.Nil(t, r)
			} else {
				require.NoError(t, err)

				require.Equal(t, tc.expected, r)
			}
		})
	}
}
