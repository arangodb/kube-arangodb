//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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

package state

import "sort"

type Server string

type Servers []Server

func (s Servers) Contains(id Server) bool {
	for _, q := range s {
		if q == id {
			return true
		}
	}

	return false
}

func (s Servers) Join(ids Servers) Servers {
	r := make(Servers, 0, len(s))

	for _, id := range ids {
		if s.Contains(id) {
			r = append(r, id)
		}
	}

	return r
}

func (s Servers) Equals(ids Servers) bool {
	if len(ids) != len(s) {
		return false
	}

	for id := range ids {
		if ids[id] != s[id] {
			return false
		}
	}

	return true
}

func (s Servers) Sort() {
	sort.Slice(s, func(i, j int) bool {
		return s[i] < s[j]
	})
}

func (s Servers) InSync(ids Servers) bool {
	if len(s) != len(ids) {
		return false
	}

	if len(s) == 0 {
		return false
	}

	if s[0] != ids[0] {
		return false
	}

	if len(s) > 1 {
		s[1:].Sort()
		ids[1:].Sort()

		if s.Equals(ids) {
			return true
		} else {
			return false
		}
	}

	return true
}
