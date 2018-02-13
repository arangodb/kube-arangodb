// Copyright (c) 2016 Pulcy.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package git

import (
	"strings"

	"github.com/coreos/go-semver/semver"
)

type TagList []string

// Gets the length of the list
func (this TagList) Len() int {
	return len(this)
}

// Swap elements at position i and j
func (this TagList) Swap(i, j int) {
	this[i], this[j] = this[j], this[i]
}

// Is element at i less than element at j?
func (this TagList) Less(i, j int) bool {

	iv := parseTag(this[i])
	jv := parseTag(this[j])

	if (iv == nil) && (jv == nil) {
		// Both non valid
		return false
	}
	if (iv != nil) && (jv == nil) {
		// this[i] is a valid, comes before non valid
		return true
	}
	if (iv == nil) && (jv != nil) {
		// this[i] is a nonvalid, comes after valid
		return false
	}

	// Sort valid versions from high to low
	return jv.LessThan(*iv)
}

// Try to parse a tag into a version.
// Returns nil when the tag cannot be parsed into a valid semver version.
func parseTag(tag string) *semver.Version {
	if strings.HasPrefix(tag, "v") {
		// Strip v prefix
		tag = tag[1:]
	}
	if strings.HasSuffix(tag, "^{}") {
		// Strip ^{} suffix (apparently added in some case by either builder of gitlab)
		// This is not considered a valid version
		return nil
	}

	v, err := semver.NewVersion(tag)
	if err != nil {
		return nil
	}
	return v
}
