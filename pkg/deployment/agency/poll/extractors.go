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

package poll

import (
	"reflect"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type ValueApplier func(out reflect.Value) error

func applyPointerCover(out reflect.Value, apply ValueApplier) error {
	if out.Kind() == reflect.Pointer {
		// DePTR object
		if out.IsZero() {
			z := reflect.New(out.Type().Elem())
			if err := applyPointerCover(z.Elem(), apply); err != nil {
				return err
			}
			out.Set(z)
		} else {
			z := out.Elem()
			if err := applyPointerCover(z, apply); err != nil {
				return err
			}
		}
		return nil
	}
	return apply(out)
}
func extract(in reflect.Value, apply ValueApplier, keys ...string) error {
	if in.Kind() != reflect.Pointer {
		return errors.Errorf("Pointer is required")
	}
	if len(keys) == 0 {
		return apply(in)
	}
	ine := in.Elem()
	switch ine.Kind() {
	case reflect.Struct:
		return extractStructValue(in, keys[0], func(out reflect.Value) error {
			if len(keys) == 1 {
				return applyPointerCover(out, apply)
			}
			return extract(out.Addr(), apply, keys[1:]...)
		})
	case reflect.Map:
		key, err := keyAsValue(ine.Type().Key(), keys[0])
		if err != nil {
			return err
		}
		return extractMapValue(in, key, func(out reflect.Value) error {
			if len(keys) == 1 {
				return apply(out)
			}
			return extract(out.Addr(), apply, keys[1:]...)
		})
	default:
		return errors.Errorf("Unknown kind %s for keys %s", ine.Kind().String(), keys)
	}
}
func remove(in reflect.Value, keys ...string) error {
	if in.Kind() != reflect.Pointer {
		return errors.Errorf("Pointer is required")
	}
	if len(keys) == 0 {
		return nil
	}
	ine := in.Elem()
	switch ine.Kind() {
	case reflect.Struct:
		if len(keys) == 1 {
			return removeStructValue(in, keys[0])
		}
		return extractStructValue(in, keys[0], func(out reflect.Value) error {
			return remove(out.Addr(), keys[1:]...)
		})
	case reflect.Map:
		key, err := keyAsValue(ine.Type().Key(), keys[0])
		if err != nil {
			return err
		}
		if len(keys) == 1 {
			// We need to remove field
			return removeMapValue(in, key)
		}
		return extractMapValue(in, key, func(out reflect.Value) error {
			return remove(out.Addr(), keys[1:]...)
		})
	default:
		return errors.Errorf("Unknown kind %s for keys %s", ine.Kind().String(), keys)
	}
}
