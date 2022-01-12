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

package v2alpha1

import (
	"time"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

// Duration is a period of time, specified in go time.Duration format.
// This is intended to allow human friendly TTL's to be specified.
type Duration string

// Validate the duration.
// Return errors when validation fails, nil on success.
func (d Duration) Validate() error {
	if d != "" {
		if _, err := time.ParseDuration(string(d)); err != nil {
			return errors.WithStack(errors.Wrapf(ValidationError, "Invalid duration: '%s': %s", string(d), err.Error()))
		}
	}
	return nil
}

// AsDuration parses the duration to a time.Duration value.
// In case of a parse error, 0 is returned.
func (d Duration) AsDuration() time.Duration {
	if d == "" {
		return 0
	}
	result, err := time.ParseDuration(string(d))
	if err != nil {
		return 0
	}
	return result
}

// NewDuration returns a reference to a Duration with given value.
func NewDuration(input Duration) *Duration {
	return &input
}

// NewDurationOrNil returns nil if input is nil, otherwise returns a clone of the given value.
func NewDurationOrNil(input *Duration) *Duration {
	if input == nil {
		return nil
	}
	return NewDuration(*input)
}

// DurationOrDefault returns the default value (or empty string) if input is nil, otherwise returns the referenced value.
func DurationOrDefault(input *Duration, defaultValue ...Duration) Duration {
	if input == nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return ""
	}
	return *input
}
