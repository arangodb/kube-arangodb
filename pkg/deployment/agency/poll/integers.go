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

func changeInteger(out *reflect.Value, diff int) error {
	var res reflect.Value
	switch out.Kind() {
	case reflect.Int:
		i, ok := out.Interface().(int)
		if !ok {
			return errors.Errorf("Type mismatch: expected int, got %s", out.Type())
		}
		res = reflect.ValueOf(i + diff)
	case reflect.Uint64:
		i, ok := out.Interface().(uint64)
		if !ok {
			return errors.Errorf("Type mismatch: expected uint64, got %s", out.Type())
		}
		res = reflect.ValueOf(i + uint64(diff))
	default:
		return errors.Errorf("Unsupported type of %s", out.Type())
	}
	out.Set(res)
	return nil
}
