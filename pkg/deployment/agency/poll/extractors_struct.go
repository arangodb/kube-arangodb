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
	goStrings "strings"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func extractStructValue(in reflect.Value, key string, apply ValueApplier) error {
	if goStrings.Contains(key, "/") {
		return errors.Errorf("Forbidden char in a key: '/'")
	}
	if in.Kind() != reflect.Pointer {
		return errors.Errorf("Pointer is required")
	}
	ine := in.Elem()
	if ine.Kind() != reflect.Struct {
		return errors.Errorf("Struct pointer is required")
	}
	for id := 0; id < ine.NumField(); id++ {
		if structFieldTaggedJSONName(ine.Type().Field(id)) == key {
			return apply(ine.Field(id))
		}
	}
	return nil
}
func removeStructValue(in reflect.Value, key string) error {
	if in.Kind() != reflect.Pointer {
		return errors.Errorf("Pointer is required")
	}
	ine := in.Elem()
	if ine.Kind() != reflect.Struct {
		return errors.Errorf("Map pointer is required")
	}
	for id := 0; id < ine.NumField(); id++ {
		if f := ine.Type().Field(id); structFieldTaggedJSONName(f) == key {
			obj := reflect.Zero(f.Type)
			ine.Field(id).Set(obj)
			return nil
		}
	}
	return nil
}
