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

func extractMapValue(in reflect.Value, key reflect.Value, apply ValueApplier) error {
	if in.Kind() != reflect.Pointer {
		return errors.Errorf("Pointer is required")
	}
	ine := in.Elem()
	if ine.Kind() != reflect.Map {
		return errors.Errorf("Map pointer is required")
	}
	if v, ok := castAsValue(key, ine.Type().Key()); !ok {
		return errors.Errorf("Invalid Map key - got %s, expected %s", key.Type(), ine.Type().Key().String())
	} else {
		key = v
	}
	// Init map if is nil
	if ine.IsNil() {
		m := reflect.MakeMap(ine.Type())
		ine.Set(m)
	}
	value := ine.MapIndex(key)
	if !value.IsValid() {
		newValue := reflect.New(ine.Type().Elem())
		value = newValue.Elem()
	}
	localModValue := reflect.New(ine.Type().Elem())
	localModValue.Elem().Set(value)
	if err := apply(localModValue.Elem()); err != nil {
		return err
	}
	ine.SetMapIndex(key, localModValue.Elem())
	return nil
}
func removeMapValue(in reflect.Value, key reflect.Value) error {
	if in.Kind() != reflect.Pointer {
		return errors.Errorf("Pointer is required")
	}
	ine := in.Elem()
	if ine.Kind() != reflect.Map {
		return errors.Errorf("Map pointer is required")
	}
	if v, ok := castAsValue(key, ine.Type().Key()); !ok {
		return errors.Errorf("Invalid Map key - got %s, expected %s", key.Type(), ine.Type().Key().String())
	} else {
		key = v
	}
	// Init map if is nil
	if ine.IsNil() {
		m := reflect.MakeMap(ine.Type())
		ine.Set(m)
	}
	value := ine.MapIndex(key)
	if !value.IsValid() {
		return nil
	}
	ine.SetMapIndex(key, reflect.Value{})
	return nil
}
