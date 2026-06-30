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

func keyAsValue(expected reflect.Type, in string) (reflect.Value, error) {
	switch expected.Kind() {
	case reflect.String:
		return reflect.ValueOf(in), nil
	default:
		return reflect.Value{}, errors.Errorf("Invalid key type")
	}
}
func castAsValue(a reflect.Value, b reflect.Type) (reflect.Value, bool) {
	if a.Type() == b {
		return a, true
	}
	if a.CanConvert(b) {
		return a.Convert(b), true
	}
	return reflect.Value{}, false
}
