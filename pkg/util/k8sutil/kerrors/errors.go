//
// DISCLAIMER
//
// Copyright 2016-2024 ArangoDB GmbH, Cologne, Germany
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

package kerrors

import (
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func Is(err error, codes ...KErrors) bool {
	if err == nil {
		return false
	}

	for _, code := range codes {
		if code.Is(err) {
			return true
		}
	}

	return false
}

type KErrors int

const (
	AlreadyExists KErrors = iota
	Conflict
	Invalid
	NotFound
	Forbidden
)

func (k KErrors) Is(err error) bool {
	switch k {
	case AlreadyExists:
		return IsAlreadyExists(err)
	case Conflict:
		return IsConflict(err)
	case Invalid:
		return IsInvalid(err)
	case NotFound:
		return IsNotFound(err)
	case Forbidden:
		return IsForbidden(err)
	default:
		return false
	}
}

func isError(err error, precondition func(err error) bool) bool {
	if err == nil {
		return false
	}

	if precondition(err) {
		return true
	}

	if c := errors.CauseWithNil(err); c == err || c == nil {
		return false
	} else {
		return isError(c, precondition)
	}
}

// IsAlreadyExists returns true if the given error is or is caused by a
// kubernetes AlreadyExistsError,
func IsAlreadyExists(err error) bool {
	return isError(err, isAlreadyExistsC)
}

func isAlreadyExistsC(err error) bool {
	return apierrors.IsAlreadyExists(err)
}

// IsConflict returns true if the given error is or is caused by a
// kubernetes ConflictError,
func IsConflict(err error) bool {
	return isError(err, isConflictC)
}

func isConflictC(err error) bool {
	return apierrors.IsConflict(err)
}

// IsForbidden returns true if the given error is or is caused by a
// kubernetes ForbiddenError,
func IsForbidden(err error) bool {
	return isError(err, isConflictC)
}

func IsForbiddenC(err error) bool {
	return apierrors.IsForbidden(err)
}

// IsNotFound returns true if the given error is or is caused by a
// kubernetes NotFoundError,
func IsNotFound(err error) bool {
	return isError(err, isNotFoundC)
}

func isNotFoundC(err error) bool {
	return apierrors.IsNotFound(err)
}

// IsInvalid returns true if the given error is or is caused by a
// kubernetes InvalidError,
func IsInvalid(err error) bool {
	return isError(err, isInvalidC)
}

func isInvalidC(err error) bool {
	return apierrors.IsInvalid(errors.Cause(err))
}

// IsForbiddenOrNotFound returns true if the given error is or is caused by a
// kubernetes NotFound or Forbidden,
func IsForbiddenOrNotFound(err error) bool {
	return isError(err, isForbiddenOrNotFoundC)
}

func isForbiddenOrNotFoundC(err error) bool {
	return apierrors.IsNotFound(err) || apierrors.IsForbidden(err)
}
