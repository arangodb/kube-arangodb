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

package helm

import (
	"encoding/json"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

func NewValues(in any) (Values, error) {
	data, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}

	return data, nil
}

type Values []byte

func (v Values) Equals(other Values) bool {
	a, err := v.Marshal()
	if err != nil {
		return false
	}

	if len(a) == 0 {
		a = nil
	}

	ad, err := json.Marshal(a)
	if err != nil {
		return false
	}

	b, err := other.Marshal()
	if err != nil {
		return false
	}

	if len(b) == 0 {
		b = nil
	}

	bd, err := json.Marshal(b)
	if err != nil {
		return false
	}

	return util.SHA256(ad) == util.SHA256(bd)
}

func (v Values) Marshal() (map[string]interface{}, error) {
	if len(v) == 0 {
		return map[string]interface{}{}, nil
	}

	var q map[string]interface{}

	if err := json.Unmarshal(v, &q); err != nil {
		return nil, err
	}

	return q, nil
}
