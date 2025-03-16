//
// DISCLAIMER
//
// Copyright 2025 ArangoDB GmbH, Cologne, Germany
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

package shared

import (
	"fmt"
	"io"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type ErrorCode int

const (
	ErrorNoError ErrorCode = iota
	ErrorUnknown
	ErrorInternal
	ErrorNotFound
	ErrorInUse
	ErrorConflict
)

func (e ErrorCode) String() string {
	switch e {
	case ErrorNoError:
		return "NoError"
	case ErrorInternal:
		return "Internal"
	case ErrorNotFound:
		return "NotFound"
	case ErrorInUse:
		return "InUse"
	case ErrorConflict:
		return "Conflict"
	default:
		return "Unknown"
	}
}

func GetError(err error) Error {
	if err == nil {
		return Error{Code: ErrorNoError}
	}

	if e, ok := getError(err); ok {
		return e
	}

	return Error{
		Code: ErrorUnknown,
	}
}

func getError(err error) (Error, bool) {
	return errors.ExtractCause[Error](err)
}

func NewError(code ErrorCode, message string, args ...interface{}) error {
	return WrapError(nil, code, message, args...)
}

func WrapError(err error, code ErrorCode, message string, args ...interface{}) error {
	return Error{
		Message: fmt.Sprintf("%s (Code %d): %s", code.String(), code, fmt.Sprintf(message, args...)),
		Code:    code,
		cause:   err,
	}
}

type Error struct {
	Message string

	Code ErrorCode

	cause error
}

func (w Error) Error() string {
	if w.cause == nil {
		return w.Message
	}

	return w.Message + ": " + w.cause.Error()
}

func (w Error) Cause() error { return w.cause }

func (w Error) Unwrap() error { return w.cause }

func (w Error) Format(s fmt.State, verb rune) {
	if w.cause == nil {
		switch verb {
		case 's', 'q', 'v':
			io.WriteString(s, w.Error())
		}

		return
	}

	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprintf(s, "%+v\n", w.Cause())
			io.WriteString(s, w.Message)
			return
		}
		fallthrough
	case 's', 'q':
		io.WriteString(s, w.Error())
	}
}
