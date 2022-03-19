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

package v2alpha1

import "encoding/json"

type ArangoTaskType string

type ArangoTaskDetails []byte

func (a ArangoTaskDetails) MarshalJSON() ([]byte, error) {
	d := make([]byte, len(a))

	copy(d, a)

	return d, nil
}

func (a *ArangoTaskDetails) UnmarshalJSON(bytes []byte) error {
	var i interface{}

	if err := json.Unmarshal(bytes, &i); err != nil {
		return err
	}

	d := make([]byte, len(bytes))

	copy(d, bytes)

	*a = d

	return nil
}

func (a ArangoTaskDetails) Get(i interface{}) error {
	return json.Unmarshal(a, &i)
}

func (a *ArangoTaskDetails) Set(i interface{}) error {
	if d, err := json.Marshal(i); err != nil {
		return err
	} else {
		*a = d
	}

	return nil
}

var _ json.Unmarshaler = &ArangoTaskDetails{}
var _ json.Marshaler = ArangoTaskDetails{}

type ArangoTaskSpec struct {
	Type ArangoTaskType `json:"type,omitempty"`

	Details ArangoTaskDetails `json:"details,omitempty"`
}
