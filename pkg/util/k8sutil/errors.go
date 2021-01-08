//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
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

package k8sutil

import (
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

// IsAlreadyExists returns true if the given error is or is caused by a
// kubernetes AlreadyExistsError,
func IsAlreadyExists(err error) bool {
	return apierrors.IsAlreadyExists(errors.Cause(err))
}

// IsConflict returns true if the given error is or is caused by a
// kubernetes ConflictError,
func IsConflict(err error) bool {
	return apierrors.IsConflict(errors.Cause(err))
}

// IsNotFound returns true if the given error is or is caused by a
// kubernetes NotFoundError,
func IsNotFound(err error) bool {
	return apierrors.IsNotFound(errors.Cause(err))
}

// IsNotFound returns true if the given error is or is caused by a
// kubernetes InvalidError,
func IsInvalid(err error) bool {
	return apierrors.IsInvalid(errors.Cause(err))
}
