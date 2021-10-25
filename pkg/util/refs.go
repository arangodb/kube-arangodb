//
// DISCLAIMER
//
// Copyright 2016-2021 ArangoDB GmbH, Cologne, Germany
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
// Author Ewout Prangsma
//

package util

import (
	"time"

	v1 "k8s.io/api/core/v1"
)

// NewString returns a reference to a string with given value.
func NewString(input string) *string {
	return &input
}

// NewStringOrNil returns nil if input is nil, otherwise returns a clone of the given value.
func NewStringOrNil(input *string) *string {
	if input == nil {
		return nil
	}
	return NewString(*input)
}

// StringOrDefault returns the default value (or empty string) if input is nil, otherwise returns the referenced value.
func StringOrDefault(input *string, defaultValue ...string) string {
	if input == nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return ""
	}
	return *input
}

// NewInt returns a reference to an int with given value.
func NewInt(input int) *int {
	return &input
}

// NewIntOrNil returns nil if input is nil, otherwise returns a clone of the given value.
func NewIntOrNil(input *int) *int {
	if input == nil {
		return nil
	}
	return NewInt(*input)
}

// IntOrDefault returns the default value (or 0) if input is nil, otherwise returns the referenced value.
func IntOrDefault(input *int, defaultValue ...int) int {
	if input == nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}
	return *input
}

// New32Int returns a reference to an int with given value.
func NewInt32(input int32) *int32 {
	return &input
}

// New32IntOrNil returns nil if input is nil, otherwise returns a clone of the given value.
func NewInt32OrNil(input *int32) *int32 {
	if input == nil {
		return nil
	}
	return NewInt32(*input)
}

// Int32OrDefault returns the default value (or 0) if input is nil, otherwise returns the referenced value.
func Int32OrDefault(input *int32, defaultValue ...int32) int32 {
	if input == nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}
	return *input
}

// NewInt64 returns a reference to an int64 with given value.
func NewInt64(input int64) *int64 {
	return &input
}

// NewInt64OrNil returns nil if input is nil, otherwise returns a clone of the given value.
func NewInt64OrNil(input *int64) *int64 {
	if input == nil {
		return nil
	}
	return NewInt64(*input)
}

// Int64OrDefault returns the default value (or 0) if input is nil, otherwise returns the referenced value.
func Int64OrDefault(input *int64, defaultValue ...int64) int64 {
	if input == nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}
	return *input
}

// NewUInt16 returns a reference to an uint16 with given value.
func NewUInt16(input uint16) *uint16 {
	return &input
}

// NewUInt16OrNil returns nil if input is nil, otherwise returns a clone of the given value.
func NewUInt16OrNil(input *uint16) *uint16 {
	if input == nil {
		return nil
	}
	return NewUInt16(*input)
}

// UInt16OrDefault returns the default value (or 0) if input is nil, otherwise returns the referenced value.
func UInt16OrDefault(input *uint16, defaultValue ...uint16) uint16 {
	if input == nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}
	return *input
}

// NewBool returns a reference to a bool with given value.
func NewBool(input bool) *bool {
	return &input
}

// NewBoolOrNil returns nil if input is nil, otherwise returns a clone of the given value.
func NewBoolOrNil(input *bool) *bool {
	if input == nil {
		return nil
	}
	return NewBool(*input)
}

// BoolOrDefault returns the default value (or false) if input is nil, otherwise returns the referenced value.
func BoolOrDefault(input *bool, defaultValue ...bool) bool {
	if input == nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return false
	}
	return *input
}

// NewDuration returns a reference to a duration with given value.
func NewDuration(input time.Duration) *time.Duration {
	return &input
}

// NewDurationOrNil returns nil if input is nil, otherwise returns a clone of the given value.
func NewDurationOrNil(input *time.Duration) *time.Duration {
	if input == nil {
		return nil
	}
	return NewDuration(*input)
}

// DurationOrDefault returns the default value (or 0) if input is nil, otherwise returns the referenced value.
func DurationOrDefault(input *time.Duration, defaultValue ...time.Duration) time.Duration {
	if input == nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}
	return *input
}

// NewPullPolicy returns a reference to a pull policy with given value.
func NewPullPolicy(input v1.PullPolicy) *v1.PullPolicy {
	return &input
}

// NewPullPolicyOrNil returns nil if input is nil, otherwise returns a clone of the given value.
func NewPullPolicyOrNil(input *v1.PullPolicy) *v1.PullPolicy {
	if input == nil {
		return nil
	}
	return NewPullPolicy(*input)
}

// PullPolicyOrDefault returns the default value (or 0) if input is nil, otherwise returns the referenced value.
func PullPolicyOrDefault(input *v1.PullPolicy, defaultValue ...v1.PullPolicy) v1.PullPolicy {
	if input == nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return ""
	}
	return *input
}
