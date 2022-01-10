//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

// Percent is a percentage between 0 and 100.
type Percent int

// Validate the given percentage.
func (p Percent) Validate() error {
	if p < 0 || p > 100 {
		return errors.WithStack(errors.Wrapf(ValidationError, "Percentage must be between 0 and 100, got %d", int(p)))
	}
	return nil
}

// NewPercent returns a reference to a percent with given value.
func NewPercent(input Percent) *Percent {
	return &input
}

// NewPercentOrNil returns nil if input is nil, otherwise returns a clone of the given value.
func NewPercentOrNil(input *Percent) *Percent {
	if input == nil {
		return nil
	}
	return NewPercent(*input)
}

// PercentOrDefault returns the default value or 0 if input is nil, otherwise returns the referenced value.
func PercentOrDefault(input *Percent, defaultValue ...Percent) Percent {
	if input == nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}
	return *input
}
