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

import "fmt"

type HashList []string

func (d HashList) Contains(hash string) bool {
	if len(d) == 0 {
		return false
	}

	for _, h := range d {
		if h == hash {
			return true
		}
	}

	return false
}

func (d HashList) ContainsSHA256(hash string) bool {
	return d.Contains(fmt.Sprintf("sha256:%s", hash))
}

func (d HashList) Equal(a HashList) bool {
	if len(d) != len(a) {
		return false
	}

	if len(d) == 0 {
		return true
	}

	for id, expected := range d {
		if a[id] != expected {
			return false
		}
	}

	return true
}
