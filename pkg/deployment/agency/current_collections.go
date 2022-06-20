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

package agency

type StateCurrentCollections map[string]StateCurrentDBCollections

func (a StateCurrentCollections) IsDBServerPresent(name Server) bool {
	for _, v := range a {
		if v.IsDBServerPresent(name) {
			return true
		}
	}

	return false
}

type StateCurrentDBCollections map[string]StateCurrentDBCollection

func (a StateCurrentDBCollections) IsDBServerPresent(name Server) bool {
	for _, v := range a {
		if v.IsDBServerPresent(name) {
			return true
		}
	}

	return false
}

type StateCurrentDBCollection map[string]StateCurrentDBShard

func (a StateCurrentDBCollection) IsDBServerPresent(name Server) bool {

	for _, v := range a {
		if v.Servers.Contains(name) {
			return true
		}
	}

	return false
}

type StateCurrentDBShard struct {
	Servers Servers `json:"servers,omitempty"`
}
