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

package types

import (
	"reflect"

	"github.com/pkg/errors"

	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
)

func EnsureTypeForwardCompatibility(a, b reflect.Type) error {
	if a.Kind() != b.Kind() {
		return errors.Errorf("Invalid kind, got %s, expected %s", a.Kind().String(), b.Kind().String())
	}

	if a == b {
		return nil
	}

	switch a.Kind() {
	case reflect.Struct:
		var err = make([]error, 0, a.NumField())

		for id := 0; id < a.NumField(); id++ {
			f := a.Field(id)

			bf, ok := b.FieldByName(f.Name)
			if !ok {
				err = append(err, shared.PrefixResourceError(f.Name, errors.Errorf("Field not defined in target")))
				continue
			}

			err = append(err, shared.PrefixResourceError(f.Name, EnsureTypeForwardCompatibility(f.Type, bf.Type)))
		}

		return shared.WithErrors(err...)
	default:
		return nil
	}
}
