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
	"fmt"
	goStrings "strings"

	"github.com/Masterminds/semver/v3"
)

// Version is an ArangoDB server version ("major.minor.patch[.sub][-suffix]"). It exists to carry a
// correct comparison that, unlike github.com/arangodb/go-driver/v2 arangodb.Version.CompareTo, does
// not fall back to a lexicographic comparison of the sub-part (which mis-orders e.g. "3.12.11-devel"
// below "3.12.8"). Cast a driver version (or any version string) with util.Version(v).
type Version string

// CompareTo returns -1 if v < other, 0 if v == other, and +1 if v > other. Every dot-separated
// numeric component is compared in order (so "3.12.12.1" > "3.12.12"), a missing component counts
// as 0 ("3.12" == "3.12.0"), and any pre-release/build suffix (e.g. "-devel", "-pre", "+meta") is
// ignored ("3.12.11-devel" == "3.12.11").
func (v Version) CompareTo(other Version) int {
	a, b := v.numbers(), other.numbers()

	n := len(a)
	if len(b) > n {
		n = len(b)
	}

	for i := 0; i < n; i++ {
		var x, y int
		if i < len(a) {
			x = a[i]
		}
		if i < len(b) {
			y = b[i]
		}

		switch {
		case x < y:
			return -1
		case x > y:
			return 1
		}
	}

	return 0
}

// Semver returns the version as a plain "major.minor.patch" string suitable for semver parsing: the
// pre-release/build suffix is dropped and any extra numeric components (the 4th in "3.12.12.1") are
// discarded, so semver constraint checks accept ArangoDB's 4-part and pre-release versions.
func (v Version) Semver() string {
	n := v.numbers()

	get := func(i int) int {
		if i < len(n) {
			return n[i]
		}
		return 0
	}

	return fmt.Sprintf("%d.%d.%d", get(0), get(1), get(2))
}

// numbers returns the dot-separated numeric components, ignoring a leading "v" and any
// pre-release/build suffix. Each component is read up to its first non-digit, so "11-devel" -> 11.
func (v Version) numbers() []int {
	s := goStrings.TrimPrefix(string(v), "v")

	if i := goStrings.IndexAny(s, "-+"); i >= 0 {
		s = s[:i]
	}

	if s == "" {
		return nil
	}

	fields := goStrings.Split(s, ".")
	out := make([]int, len(fields))
	for i, f := range fields {
		out[i] = leadingInt(f)
	}
	return out
}

func leadingInt(s string) int {
	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			break
		}
		n = n*10 + int(r-'0')
	}
	return n
}

// VersionConstrain is a semver constraint expression (e.g. ">= 3.12.0 < 3.13.0"). Cast a string with
// util.VersionConstrain(s); Validate reports any parse error.
type VersionConstrain string

// Empty reports whether the constraint is unset, i.e. imposes no version requirement.
func (c VersionConstrain) Empty() bool {
	return c == ""
}

// Validate reports whether the given version satisfies the constraint. An empty constraint imposes no
// requirement and always returns true. Any pre-release suffix on the version is dropped before
// checking, so a "-devel" build is treated as its release version.
func (c VersionConstrain) Validate(version Version) (bool, error) {
	if c.Empty() {
		return true, nil
	}

	constrain, err := semver.NewConstraint(string(c))
	if err != nil {
		return false, err
	}

	// Normalise to major.minor.patch: drops the pre-release/build suffix and any extra numeric
	// component (e.g. the 4th in "3.12.12.1"), so ArangoDB's 4-part and "-devel" versions parse.
	ver, err := semver.NewVersion(version.Semver())
	if err != nil {
		return false, err
	}

	return constrain.Check(ver), nil
}
