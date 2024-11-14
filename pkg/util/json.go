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

package util

import (
	"encoding/json"

	"sigs.k8s.io/yaml"
)

func JSONRemarshal[A, B any](in A) (B, error) {
	d, err := json.Marshal(in)
	if err != nil {
		return Default[B](), err
	}

	var o B

	if err := json.Unmarshal(d, &o); err != nil {
		return Default[B](), err
	}

	return o, nil
}

func JsonOrYamlUnmarshal[T any](b []byte) (T, error) {
	var z T

	if json.Valid(b) {
		if err := json.Unmarshal(b, &z); err != nil {
			return Default[T](), err
		}

		return z, nil
	}

	if err := yaml.UnmarshalStrict(b, &z); err != nil {
		return Default[T](), err
	}

	return z, nil
}
