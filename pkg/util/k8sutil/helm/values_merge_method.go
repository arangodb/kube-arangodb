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

package helm

import "github.com/arangodb/kube-arangodb/pkg/util/errors"

type ValuesMergeMethod int

const (
	MergeMaps ValuesMergeMethod = 1 << iota
)

func (v ValuesMergeMethod) Has(o ValuesMergeMethod) bool {
	return v&o == o
}

func (v ValuesMergeMethod) Merge(a, b Values) (Values, error) {
	av, err := a.Marshal()
	if err != nil {
		return nil, err
	}

	bv, err := b.Marshal()
	if err != nil {
		return nil, err
	}

	z, err := v.mergeMaps(av, bv)
	if err != nil {
		return nil, err
	}

	return NewValues(z)
}

func (v ValuesMergeMethod) mergeMaps(a, b map[string]interface{}) (map[string]interface{}, error) {
	if len(a) == 0 && len(b) == 0 {
		return map[string]interface{}{}, nil
	}
	if len(a) == 0 {
		return b, nil
	}

	if len(b) == 0 {
		return a, nil
	}

	ret := make(map[string]interface{})

	for k, v := range a {
		ret[k] = v
	}

	for k, o := range b {
		z, ok := ret[k]
		if !ok {
			ret[k] = o
			continue
		}

		if v.Has(MergeMaps) {
			if z != nil && o != nil {
				av, aok := z.(map[string]interface{})
				bv, bok := o.(map[string]interface{})

				if aok || bok {
					if !aok || !bok {
						return nil, errors.Errorf("Invalid types during Map merge")
					}

					rv, err := v.mergeMaps(av, bv)
					if err != nil {
						return nil, err
					}

					ret[k] = rv
					continue
				}
			}
		}

		// Override at the end
		ret[k] = o
	}

	return ret, nil
}
