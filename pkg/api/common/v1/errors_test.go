//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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

package v1

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommonErrorsWrapping(t *testing.T) {
	errors := []struct {
		name           string
		errorFunc      func(string, ...interface{}) error
		validationFunc func(error) bool
	}{

		{"Canceled", Canceled, IsCanceled},
		{"InvalidArgument", InvalidArgument, IsInvalidArgument},
		{"NotFound", NotFound, IsNotFound},
		{"AlreadyExists", AlreadyExists, IsAlreadyExists},
		{"PreconditionFailed", PreconditionFailed, IsPreconditionFailed},
		{"Unavailable", Unavailable, IsUnavailable},
	}
	for idx, testCase := range errors {
		t.Run(testCase.name, func(t *testing.T) {
			e := testCase.errorFunc("%s error", testCase.name)
			wrapped := fmt.Errorf("Wraps: %w", e)
			wrapped2 := fmt.Errorf("Wraps another one: %w", wrapped)
			for idx2 := range errors {
				assert.Equal(t, idx == idx2, errors[idx2].validationFunc(e))
				assert.Equal(t, idx == idx2, errors[idx2].validationFunc(wrapped), "wrapped error %s is not detected as an error", testCase.name)
				assert.Equal(t, idx == idx2, errors[idx2].validationFunc(wrapped2), "wrapped error %s is not detected as an error", testCase.name)
			}
		})
	}
}
