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

package panics

import (
	"time"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/version"
)

type PanicErrorDetails struct {
	OperatorVersion version.InfoV1

	Time time.Time
}

type PanicError interface {
	error

	Details() PanicErrorDetails
	PanicCause() interface{}
	Stack() StackEntries
}

func IsPanicError(err error) (PanicError, bool) {
	if err == nil {
		return nil, false
	}

	if e, ok := err.(PanicError); ok {
		return e, true
	}

	return IsPanicError(errors.CauseWithNil(err))
}

func newPanicError(cause interface{}, stack StackEntries) PanicError {
	return panicError{
		cause: cause,
		stack: stack,
		details: PanicErrorDetails{
			OperatorVersion: version.GetVersionV1(),
			Time:            time.Now(),
		},
	}
}

type panicError struct {
	cause interface{}

	stack   StackEntries
	details PanicErrorDetails
}

func (p panicError) Details() PanicErrorDetails {
	return p.details
}

func (p panicError) PanicCause() interface{} {
	return p.cause
}

func (p panicError) Stack() StackEntries {
	return p.stack
}

func (p panicError) Error() string {
	return "Panic Received"
}
