//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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

package state

import (
	"crypto/sha256"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

type Exists []byte

func (d Exists) Hash() string {
	if d == nil {
		return ""
	}

	return util.SHA256(d)
}

func (d Exists) Exists() bool {
	return d != nil
}

func (d *Exists) UnmarshalJSON(bytes []byte) error {
	if bytes == nil {
		*d = nil
		return nil
	}

	data := sha256.Sum256(bytes)
	allData := data[:]

	*d = allData
	return nil
}
