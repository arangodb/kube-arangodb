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

package util

// NewType returns a reference to a simple type with given value.
func NewType[T interface{}](input T) *T {
	return &input
}

// NewTypeOrNil returns nil if input is nil, otherwise returns a clone of the given value.
func NewTypeOrNil[T interface{}](input *T) *T {
	if input == nil {
		return nil
	}

	return NewType(*input)
}

// TypeOrDefault returns the default value (or T default value) if input is nil, otherwise returns the referenced value.
func TypeOrDefault[T interface{}](input *T, defaultValue ...T) T {
	if input == nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		var def T
		return def
	}
	return *input
}

// BoolSwitch define bool switch for defined types - in case of true t T is returned, in case of false f T
func BoolSwitch[T interface{}](s bool, t, f T) T {
	if s {
		return t
	}

	return f
}
