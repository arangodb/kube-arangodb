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

package util

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_VersionConstrain(t *testing.T) {
	type constrain struct {
		version string
		valid   bool
	}

	validate := func(version string, checks ...constrain) {
		t.Run(version, func(t *testing.T) {
			vc, err := NewVersionConstrain(version)
			require.NoError(t, err)

			for _, el := range checks {
				t.Run(el.version, func(t *testing.T) {
					b, err := vc.Validate(el.version)
					require.NoError(t, err)

					if el.valid {
						require.True(t, b)
					} else {
						require.False(t, b)
					}
				})
			}
		})
	}

	validate(">= 1.2.3 < 1.3.0",
		constrain{
			version: "1.2.3",
			valid:   true,
		},
		constrain{
			version: "1.2.5",
			valid:   true,
		},
		constrain{
			version: "1.3.0",
		},
		constrain{
			version: "v1.2.3-abcdefg",
			valid:   true,
		},
	)

	validate(">= 1.2.3 < 1.3.0 || >= 1.3.1 < 1.4.0",
		constrain{
			version: "1.2.3",
			valid:   true,
		},
		constrain{
			version: "1.2.5",
			valid:   true,
		},
		constrain{
			version: "1.3.0",
		},
		constrain{
			version: "v1.2.3-abcdefg",
			valid:   true,
		},
		constrain{
			version: "1.3.5",
			valid:   true,
		},
		constrain{
			version: "1.4.0",
		},
	)

	validate("~ 1",
		constrain{
			version: "1.2.3",
			valid:   true,
		},
		constrain{
			version: "1.2.5",
			valid:   true,
		},
		constrain{
			version: "1.3.0",
			valid:   true,
		},
		constrain{
			version: "v1.2.3-abcdefg",
			valid:   true,
		},
		constrain{
			version: "1.3.5",
			valid:   true,
		},
		constrain{
			version: "1.4.0",
			valid:   true,
		},
		constrain{
			version: "2.0.0",
		},
	)
}
