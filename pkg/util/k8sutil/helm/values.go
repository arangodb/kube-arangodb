//
// DISCLAIMER
//
// Copyright 2024-2025 ArangoDB GmbH, Cologne, Germany
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

func NewMergeValues(opts ValuesMergeMethod, vs ...any) (Values, error) {
	if len(vs) == 0 {
		return nil, nil
	}

	o, err := NewValues(vs[0])
	if err != nil {
		return nil, err
	}

	if len(vs) == 1 {
		return o, nil
	}

	for _, el := range vs[1:] {
		a, err := NewValues(el)
		if err != nil {
			return nil, err
		}

		no, err := opts.Merge(o, a)
		if err != nil {
			return nil, err
		}

		o = no
	}

	return o, nil
}

func NewValues(in any) (Values, error) {
	data, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}

	return data, nil
}

type Values []byte

func (v Values) Equals(other Values) bool {
	a, err := v.MarshalJSON()
	if err != nil {
		return false
	}

	b, err := other.MarshalJSON()
	if err != nil {
		return false
	}

	return util.SHA256(a) == util.SHA256(b)
}

func (v *Values) UnmarshalJSON(in []byte) error {
	q := make(Values, len(in))

	copy(q, in)

	*v = q

	return nil
}

func (v Values) MarshalJSON() ([]byte, error) {
	return v, nil
}

func (v Values) String() string {
	return string(v)
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
