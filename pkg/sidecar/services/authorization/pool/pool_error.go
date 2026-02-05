//
// DISCLAIMER
//
// Copyright 2026 ArangoDB GmbH, Cologne, Germany
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

package pool

import (
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type PoolOutOfBoundsError struct{}

func (PoolOutOfBoundsError) Error() string {
	return "pool out of bounds"
}

func IsPoolOutOfBoundsError(err error) bool {
	var poolOutOfBoundsError PoolOutOfBoundsError
	ok := errors.As(err, &poolOutOfBoundsError)
	return ok
}

type PoolNoChangeError struct{}

func (PoolNoChangeError) Error() string { return "no change error" }

func IsPoolNoChangeError(err error) bool {
	var poolNoChangeError PoolNoChangeError
	ok := errors.As(err, &poolNoChangeError)
	return ok
}

type PoolAlreadyExists struct{}

func (PoolAlreadyExists) Error() string { return "already exists" }

func IsPoolAlreadyExistsError(err error) bool {
	var poolAlreadyExistsError PoolAlreadyExists
	ok := errors.As(err, &poolAlreadyExistsError)
	return ok
}

type PoolNotFound struct{}

func (PoolNotFound) Error() string { return "not found" }

func IsPoolNotFound(err error) bool {
	var poolNotFound PoolNotFound
	ok := errors.As(err, &poolNotFound)
	return ok
}
