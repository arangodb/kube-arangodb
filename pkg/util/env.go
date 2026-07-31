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

package util

import (
	"os"
	"strconv"
	goStrings "strings"
	"time"
)

// EnvironmentVariable is a wrapper to get environment variables
type EnvironmentVariable string

// String return string representation of environment variable name
func (e EnvironmentVariable) String() string {
	return string(e)
}

// Lookup for environment variable
func (e EnvironmentVariable) Lookup() (string, bool) {
	return os.LookupEnv(e.String())
}

// Exists check if variable is defined
func (e EnvironmentVariable) Exists() bool {
	_, exists := e.Lookup()
	return exists
}

// Get fetch variable. If variable does not exist empty string is returned
func (e EnvironmentVariable) Get() string {
	value, _ := e.Lookup()
	return value
}

// GetOrDefault fetch variable. If variable is not defined default value is returned
func (e EnvironmentVariable) GetOrDefault(d string) string {
	if value, exists := e.Lookup(); exists {
		return value
	}

	return d
}

// NormalizeEnv normalizes environment variables.
func NormalizeEnv(env string) string {
	r := goStrings.NewReplacer(".", "_", "-", "_")
	return goStrings.ToUpper(r.Replace(env))
}

// EnvironmentVariableTyped is a typed wrapper around an environment variable name.
// T determines the parse strategy via the standard library:
//   - int: strconv.Atoi
//   - time.Duration: time.ParseDuration
type EnvironmentVariableTyped[T any] string

// String returns the environment variable name.
func (e EnvironmentVariableTyped[T]) String() string {
	return string(e)
}

// Lookup returns the parsed value and whether the variable exists.
func (e EnvironmentVariableTyped[T]) Lookup() (T, bool, error) {
	raw, ok := os.LookupEnv(string(e))
	if !ok {
		var zero T
		return zero, false, nil
	}

	v, err := parseEnvValue[T](raw)
	return v, true, err
}

// Get returns the parsed value or the zero value if not set.
func (e EnvironmentVariableTyped[T]) Get() (T, error) {
	v, _, err := e.Lookup()
	return v, err
}

// GetOrDefault returns the parsed value or the given default if not set or on parse error.
func (e EnvironmentVariableTyped[T]) GetOrDefault(d T) T {
	v, ok, err := e.Lookup()
	if !ok || err != nil {
		return d
	}
	return v
}

func parseEnvValue[T any](raw string) (T, error) {
	var zero T
	switch p := any(&zero).(type) {
	case *int:
		v, err := strconv.Atoi(raw)
		if err != nil {
			return zero, err
		}
		*p = v
	case *time.Duration:
		v, err := time.ParseDuration(raw)
		if err != nil {
			return zero, err
		}
		*p = v
	case *string:
		*p = raw
	case *bool:
		v, err := strconv.ParseBool(raw)
		if err != nil {
			return zero, err
		}
		*p = v
	}
	return zero, nil
}
