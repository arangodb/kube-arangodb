//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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

// KeyValuePairList is a strong-typed list of KeyValuePair
type KeyValuePairList []*KeyValuePair

// GetValue gets the value for the requested key or nil if it doesn't exist
func (list KeyValuePairList) GetValue(key string) *string {
	for _, kv := range list {
		if kv.GetKey() == key {
			v := kv.GetValue()
			return &v
		}
	}
	return nil
}

// UpsertPair update or insert the given value for the requested key
// Returns inserted (otherwise updated)
func (list *KeyValuePairList) UpsertPair(key, value string) bool {
	if list == nil {
		return false
	}
	for _, kv := range *list {
		if kv.GetKey() == key {
			kv.Value = value
			return false
		}
	}
	*list = append(*list, &KeyValuePair{
		Key:   key,
		Value: value,
	})
	return true
}
