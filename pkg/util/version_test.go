//
// DISCLAIMER
//
// Copyright 2024-2026 ArangoDB GmbH, Cologne, Germany
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
			vc := VersionConstrain(version)

			for _, el := range checks {
				t.Run(el.version, func(t *testing.T) {
					b, err := vc.Validate(Version(el.version))
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

	// ArangoDB-style versions: 4-part ("3.12.12.1") and pre-release/build suffixes must be accepted
	// by Validate even though the Masterminds semver parser only understands major.minor.patch.
	validate(">= 3.12.0 < 3.13.0",
		constrain{version: "3.12.0", valid: true},
		constrain{version: "3.12.11", valid: true},
		constrain{version: "3.12.12.1", valid: true},     // 4-part normalises to 3.12.12
		constrain{version: "3.12.11-devel", valid: true}, // pre-release dropped -> 3.12.11
		constrain{version: "3.12.11-pre", valid: true},
		constrain{version: "3.12.0+meta", valid: true},
		constrain{version: "3.11.99"},
		constrain{version: "3.13.0"},
		constrain{version: "3.13.0-devel"}, // -> 3.13.0, not < 3.13.0
		constrain{version: "4.0.0.1"},
	)
}

func Test_Version_CompareTo(t *testing.T) {
	type c struct {
		a, b Version
		want int
	}

	for _, tc := range []c{
		// equal
		{"3.12.11", "3.12.11", 0},
		{"3.12", "3.12.0", 0}, // missing component counts as 0
		// major / minor
		{"4.0.0", "3.99.99", 1},
		{"3.11.0", "3.12.0", -1},
		// numeric patch - the driver bug: must NOT be lexicographic
		{"3.12.11", "3.12.8", 1},
		{"3.12.8", "3.12.11", -1},
		{"3.12.10", "3.12.9", 1},
		// 4-part
		{"3.12.12.1", "3.12.12", 1},
		{"3.12.12", "3.12.12.1", -1},
		{"3.12.12.1", "3.12.12.1", 0},
		{"3.12.12.1", "3.12.12.2", -1},
		// pre-release / build suffix ignored
		{"3.12.11-devel", "3.12.11", 0},
		{"3.12.11-devel", "3.12.8", 1},
		{"3.12.8", "3.12.11-devel", -1},
		{"3.12.12.1-devel", "3.12.12.1", 0},
		{"3.12.11-pre", "3.12.11-rc1", 0},
		{"v3.12.11", "3.12.11", 0},
		{"3.12.11+build", "3.12.11", 0},
	} {
		t.Run(string(tc.a)+"_vs_"+string(tc.b), func(t *testing.T) {
			require.Equal(t, tc.want, tc.a.CompareTo(tc.b))
			// Comparison must be antisymmetric.
			require.Equal(t, -tc.want, tc.b.CompareTo(tc.a))
		})
	}
}
