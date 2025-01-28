//
// DISCLAIMER
//
// Copyright 2023-2025 ArangoDB GmbH, Cologne, Germany
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

package operator

import (
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func Temporary(cause error, msg string, args ...interface{}) error {
	if cause == nil {
		return temporary{
			cause: errors.Errorf(msg, args...),
		}
	}

	return temporary{
		cause: errors.Wrapf(cause, msg, args...),
	}
}

type temporary struct {
	cause error
}

func (t temporary) Error() string {
	return t.cause.Error()
}

func (t temporary) Temporary() bool {
	return true
}

func (t temporary) Cause() error {
	return t.cause
}

func IsTemporary(err error) bool {
	if _, ok := errors.ExtractCause[temporary](err); ok {
		return true
	}

	return false
}
