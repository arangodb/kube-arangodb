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

package util

import (
	"os"
	"strings"
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
	r := strings.NewReplacer(".", "_", "-", "_")
	return strings.ToUpper(r.Replace(env))
}
