//
// DISCLAIMER
//
// Copyright 2016-2025 ArangoDB GmbH, Cologne, Germany
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

type Hash interface {
	Hash() string
}

func CompareJSON[T interface{}](a, b T) (bool, error) {
	ad, err := SHA256FromJSON(a)
	if err != nil {
		return false, err
	}
	bd, err := SHA256FromJSON(b)
	if err != nil {
		return false, err
	}

	return ad == bd, nil
}

func CompareJSONP[T interface{}](a, b *T) (bool, error) {
	var a1, b1 T

	if a != nil {
		a1 = *a
	}

	if b != nil {
		b1 = *b
	}

	return CompareJSON(a1, b1)
}
