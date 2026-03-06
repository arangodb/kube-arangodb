//
// DISCLAIMER
//
// Copyright 2016-2026 ArangoDB GmbH, Cologne, Germany
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

package patch

import (
	"encoding/json"

	jsonpatch "github.com/evanphx/json-patch"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

func NewPatch(items ...Item) Patch {
	return items
}

func Object[T any](in T, patch Patch) (T, error) {
	data, err := json.Marshal(in)
	if err != nil {
		return util.Default[T](), err
	}

	ndata, err := patch.Apply(data)

	if err != nil {
		return util.Default[T](), err
	}

	return util.Unmarshal[T](ndata)
}

type Patch []Item

func (p Patch) Apply(in []byte) ([]byte, error) {
	if len(p) == 0 {
		return in, nil
	}

	patch, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}

	z, err := jsonpatch.DecodePatch(patch)
	if err != nil {
		return nil, err
	}

	patched, err := z.Apply(in)
	if err != nil {
		return nil, err
	}

	return patched, nil
}

func (p Patch) Add(items ...Item) Patch {
	q := make(Patch, len(p)+len(items))
	copy(q, p)
	copy(q[len(p):], items)
	return q
}

func (p Patch) ItemAdd(path Path, value interface{}) Patch {
	return p.Add(ItemAdd(path, value))
}

func (p Patch) ItemReplace(path Path, value interface{}) Patch {
	return p.Add(ItemReplace(path, value))
}

func (p Patch) ItemRemove(path Path) Patch {
	return p.Add(ItemRemove(path))
}

func (p Patch) Marshal() ([]byte, error) {
	return json.Marshal(p)
}
