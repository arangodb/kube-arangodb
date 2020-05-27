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

package client

type EncryptionKeyEntry struct {
	Sha string `json:"sha256,omitempty"`
}

type EncryptionDetailsResult struct {
	Keys []EncryptionKeyEntry `json:"encryption-keys,omitempty"`
}

func (e EncryptionDetailsResult) KeysPresent(m map[string][]byte) bool {
	if len(e.Keys) != len(m) {
		return false
	}

	for key := range m {
		ok := false
		for _, entry := range e.Keys {
			if entry.Sha == key {
				ok = true
				break
			}
		}
		if !ok {
			return false
		}
	}

	return true
}

type EncryptionDetails struct {
	Result EncryptionDetailsResult `json:"result,omitempty"`
}
