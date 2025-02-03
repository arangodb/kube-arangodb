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

import (
	"reflect"

	"github.com/pkg/errors"
)

// NewPointer returns a reference to a copy of the pointer value
func NewPointer[T any](input *T) *T {
	if input == nil {
		return nil
	}

	return NewType(*input)
}

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

// OptionalType returns the default value (or T default value)
func OptionalType[T any](in *T, v T) T {
	if in == nil {
		return v
	}

	return *in
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

// WithDefault returns generic default value for type T if in is nil
func WithDefault[T interface{}](in *T) T {
	if in == nil {
		return Default[T]()
	}

	return *in
}

// IsDefault returns if object is default
func IsDefault[T any](in T) bool {
	t := reflect.ValueOf(in)

	if t.IsValid() {
		return false
	}

	return t.IsZero() || t.IsNil()
}

// Default returns generic default value for type T
func Default[T interface{}]() T {
	var d T
	return d
}

// DefaultInterface returns generic default value for type T as Interface
func DefaultInterface[T interface{}]() interface{} {
	var d T
	return d
}

// First returns first not nil value
func First[T interface{}](input ...*T) *T {
	for _, i := range input {
		if i != nil {
			return i
		}
	}

	return nil
}

// FirstNotDefault returns first not default value
func FirstNotDefault[T interface{}](input ...T) T {
	for _, i := range input {
		if !reflect.DeepEqual(i, Default[T]()) {
			return i
		}
	}

	return Default[T]()
}

// LastFromList returns last element on the list
func LastFromList[T interface{}](in []T) T {
	return in[len(in)-1]
}

// BoolSwitch define bool switch for defined types - in case of true t T is returned, in case of false f T
func BoolSwitch[T interface{}](s bool, t, f T) T {
	if s {
		return t
	}

	return f
}

// InitType initialise object if it is nil pointer
func InitType[T interface{}](in *T) *T {
	if in != nil {
		return in
	}

	var q T
	return &q
}

func DeepType[T any]() (T, error) {
	var z T

	if err := InitDeepType(&z); err != nil {
		return Default[T](), err
	}

	return z, nil
}

func InitDeepType(in any) error {
	return initDeepType(reflect.ValueOf(in))
}

func initDeepType(v reflect.Value) error {
	switch v.Kind() {
	case reflect.Pointer:
		if !v.Elem().CanSet() {
			return errors.Errorf("Unable to set interface")
		}

		switch v.Elem().Kind() {
		case reflect.Pointer:
			nv := reflect.New(v.Type().Elem().Elem())
			if err := initDeepType(nv); err != nil {
				return err
			}
			v.Elem().Set(nv)
		default:
			v.Elem().Set(reflect.New(v.Elem().Type()).Elem())
		}
	}

	return nil
}

type ConditionalFunction[T interface{}] func() (T, bool)
type ConditionalP1Function[T, P1 interface{}] func(p1 P1) (T, bool)

func CheckConditionalNil[T interface{}](in ConditionalFunction[T]) *T {
	if v, ok := in(); ok {
		return &v
	}

	return nil
}

func CheckConditionalP1Nil[T, P1 interface{}](in ConditionalP1Function[T, P1], p1 P1) *T {
	if v, ok := in(p1); ok {
		return &v
	}

	return nil
}

func Or(in ...bool) bool {
	for _, v := range in {
		if v {
			return true
		}
	}

	return false
}

func And(in ...bool) bool {
	for _, v := range in {
		if !v {
			return false
		}
	}

	return len(in) > 0
}

func InitOptional[T any](in *T, ok bool) *T {
	if ok {
		return in
	}

	var z T
	return &z
}
