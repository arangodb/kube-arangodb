//
// DISCLAIMER
//
// Copyright 2018 ArangoDB GmbH, Cologne, Germany
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

import "time"

// String returns a reference to a string with given value.
func String(input string) *string {
	return &input
}

// StringOrNil returns nil if input is nil, otherwise returns a clone of the given value.
func StringOrNil(input *string) *string {
	if input == nil {
		return nil
	}
	return String(*input)
}

// Int returns a reference to an int with given value.
func Int(input int) *int {
	return &input
}

// IntOrNil returns nil if input is nil, otherwise returns a clone of the given value.
func IntOrNil(input *int) *int {
	if input == nil {
		return nil
	}
	return Int(*input)
}

// Bool returns a reference to a bool with given value.
func Bool(input bool) *bool {
	return &input
}

// BoolOrNil returns nil if input is nil, otherwise returns a clone of the given value.
func BoolOrNil(input *bool) *bool {
	if input == nil {
		return nil
	}
	return Bool(*input)
}

// Duration returns a reference to a duration with given value.
func Duration(input time.Duration) *time.Duration {
	return &input
}

// DurationOrNil returns nil if input is nil, otherwise returns a clone of the given value.
func DurationOrNil(input *time.Duration) *time.Duration {
	if input == nil {
		return nil
	}
	return Duration(*input)
}
