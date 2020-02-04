//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
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
// Author Adam Janikowski
//

package pod

import "strings"

// OptionPair key value pair builder
type OptionPair struct {
	Key   string
	Value string
}

// CompareTo returns -1 if o < other, 0 if o == other, 1 otherwise
func (o OptionPair) CompareTo(other OptionPair) int {
	rc := strings.Compare(o.Key, other.Key)
	if rc < 0 {
		return -1
	} else if rc > 0 {
		return 1
	}
	return strings.Compare(o.Value, other.Value)
}

func NewOptionPair(pairs ...OptionPair) []OptionPair {
	return pairs
}
