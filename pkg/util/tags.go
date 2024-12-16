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
	"reflect"

	"github.com/pkg/errors"
)

func ExtractTags[OUT any](t reflect.StructField) (OUT, error) {
	var out OUT

	v := reflect.ValueOf(&out).Elem()
	z := v.Type()

	if z.Kind() != reflect.Struct {
		return Default[OUT](), errors.Errorf("Only Struct kind allowed")
	}

	for id := 0; id < z.NumField(); id++ {
		f := z.Field(id)
		vf := v.Field(id)

		if !f.IsExported() {
			continue
		}

		if f.Anonymous {
			continue
		}

		tag, ok := f.Tag.Lookup("tag")
		if !ok {
			continue
		}

		if f.Type != reflect.TypeOf(Default[*string]()) {
			return Default[OUT](), errors.Errorf("Tagged fields can be only *string type")
		}

		if v, ok := t.Tag.Lookup(tag); ok {
			vf.Set(reflect.ValueOf(&v))
		} else {
			vf.Set(reflect.Zero(vf.Type()))
		}
	}

	return out, nil
}
