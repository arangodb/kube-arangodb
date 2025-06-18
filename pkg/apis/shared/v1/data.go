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

package v1

import (
	"encoding/base64"
	"encoding/json"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

var _ json.Marshaler = &Data{}
var _ json.Unmarshaler = &Data{}

func NewData[T any](in T) (Data, error) {
	return json.Marshal(in)
}

func FromData[T any](in Data) (T, error) {
	var z T
	if err := json.Unmarshal(in, &z); err != nil {
		return z, err
	}

	return z, nil
}

// Data keeps the representation of the object in the base64 encoded string
type Data []byte

func (d Data) MarshalJSON() ([]byte, error) {
	s := base64.StdEncoding.EncodeToString(d)

	return json.Marshal(s)
}

func (d Data) IsZero() bool {
	return len(d) == 0
}

func (d Data) SHA256() string {
	return util.SHA256(d)
}

func (d *Data) UnmarshalJSON(bytes []byte) error {
	if d == nil {
		return errors.Errorf("nil object provided")
	}

	var s string

	if err := json.Unmarshal(bytes, &s); err != nil {
		return err
	}

	ret, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return err
	}

	*d = ret

	return nil
}
