//
// DISCLAIMER
//
// Copyright 2025 ArangoDB GmbH, Cologne, Germany
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

import (
	"encoding/json"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

var _ json.Marshaler = &Any{}
var _ json.Unmarshaler = &Any{}

type Any []byte

func NewAny[T any](in T) (Any, error) {
	return json.Marshal(in)
}

func FromAny[T any](in Any) (T, error) {
	var z T
	if err := json.Unmarshal(in, &z); err != nil {
		return z, err
	}

	return z, nil
}

func (d Any) MarshalJSON() ([]byte, error) {
	ret := make([]byte, len(d))
	copy(ret, d)
	return ret, nil
}

func (d Any) IsZero() bool {
	return len(d) == 0 || string(d) == "{}"
}

func (Any) OpenAPISchemaType() []string { return []string{"object"} }

func (Any) OpenAPIXPreserveUnknownFields() bool {
	return true
}

func (d Any) SHA256() string {
	return util.SHA256(d)
}

func (d Any) Equals(o Any) bool {
	return d.SHA256() == o.SHA256()
}

func (d *Any) UnmarshalJSON(bytes []byte) error {
	if d == nil {
		return errors.Errorf("nil object provided")
	}

	ret := make([]byte, len(bytes))
	copy(ret, bytes)

	*d = ret

	return nil
}
