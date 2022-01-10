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

// Environment in which to run the cluster
type Environment string

const (
	// EnvironmentDevelopment yields a cluster optimized for development
	EnvironmentDevelopment Environment = "Development"
	// EnvironmentProduction yields a cluster optimized for production
	EnvironmentProduction Environment = "Production"
)

// Validate the environment.
// Return errors when validation fails, nil on success.
func (e Environment) Validate() error {
	switch e {
	case EnvironmentDevelopment, EnvironmentProduction:
		return nil
	default:
		return errors.WithStack(errors.Wrapf(ValidationError, "Unknown environment: '%s'", string(e)))
	}
}

// IsProduction returns true when the given environment is a production environment.
func (e Environment) IsProduction() bool {
	return e == EnvironmentProduction
}

// NewEnvironment returns a reference to a string with given value.
func NewEnvironment(input Environment) *Environment {
	return &input
}

// NewEnvironmentOrNil returns nil if input is nil, otherwise returns a clone of the given value.
func NewEnvironmentOrNil(input *Environment) *Environment {
	if input == nil {
		return nil
	}
	return NewEnvironment(*input)
}

// EnvironmentOrDefault returns the default value (or empty string) if input is nil, otherwise returns the referenced value.
func EnvironmentOrDefault(input *Environment, defaultValue ...Environment) Environment {
	if input == nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return ""
	}
	return *input
}
