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

package v1

import (
	"encoding/base64"
	"encoding/json"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

var _ json.Marshaler = &Data{}
var _ json.Unmarshaler = &Data{}

type Data []byte

func (d Data) MarshalJSON() ([]byte, error) {
	s := base64.StdEncoding.EncodeToString(d)

	return json.Marshal(s)
}

func (d Data) SHA256() string {
	return util.SHA256(d)
}

func (d *Data) UnmarshalJSON(bytes []byte) error {
	if d == nil {
		return errors.Errorf("nil object provided")
	}

	ret, err := base64.StdEncoding.DecodeString(string(bytes))
	if err != nil {
		return err
	}

	*d = ret

	return nil
}
