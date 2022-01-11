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

import "github.com/arangodb/kube-arangodb/pkg/util/errors"

var (
	// ValidationError indicates a validation failure
	ValidationError = errors.New("validation failed")

	// AlreadyExistsError indicates an object that already exists
	AlreadyExistsError = errors.New("already exists")

	// NotFoundError indicates an object that cannot be found
	NotFoundError = errors.New("not found")
)

// IsValidation return true when the given error is or is caused by a ValidationError.
func IsValidation(err error) bool {
	return errors.Cause(err) == ValidationError
}

// IsAlreadyExists return true when the given error is or is caused by a AlreadyExistsError.
func IsAlreadyExists(err error) bool {
	return errors.Cause(err) == AlreadyExistsError
}

// IsNotFound return true when the given error is or is caused by a NotFoundError.
func IsNotFound(err error) bool {
	return errors.Cause(err) == NotFoundError
}
