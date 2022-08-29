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

package retry

import (
	"context"
	"time"

	"github.com/cenkalti/backoff"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type permanentError struct {
	Err error
}

func (e *permanentError) Error() string {
	return e.Err.Error()
}

func (e *permanentError) Cause() error {
	return e.Err
}

// Permanent makes the given error a permanent failure
// that stops the Retry loop immediately.
func Permanent(err error) error {
	return &permanentError{Err: err}
}

func isPermanent(err error) (*permanentError, bool) {
	for err != nil {
		if pe, ok := err.(*permanentError); ok {
			return pe, true
		}
		cause, ok := err.(errors.Causer)
		if !ok {
			break
		}
		err = cause.Cause()
	}
	return nil, false
}

// retry the given operation until it succeeds,
// has a permanent failure or times out.
func retry(ctx context.Context, op func() error, timeout time.Duration) error {
	var failure error
	wrappedOp := func() error {
		if err := op(); err == nil {
			return nil
		} else {
			if pe, ok := isPermanent(err); ok {
				// Detected permanent error
				failure = pe.Err
				return nil
			} else {
				return err
			}
		}
	}

	eb := backoff.NewExponentialBackOff()
	eb.MaxElapsedTime = timeout
	eb.MaxInterval = timeout / 3

	var b backoff.BackOff
	if ctx != nil {
		b = backoff.WithContext(eb, ctx)
	} else {
		b = eb
	}

	if err := backoff.Retry(wrappedOp, b); err != nil {
		return errors.WithStack(err)
	}
	if failure != nil {
		return errors.WithStack(failure)
	}
	return nil
}

// Retry the given operation until it succeeds,
// has a permanent failure or times out.
func Retry(op func() error, timeout time.Duration) error {
	return retry(nil, op, timeout) // nolint:staticcheck
}

// RetryWithContext retries the given operation until it succeeds,
// has a permanent failure or times out.
// The timeout is the minimum between the timeout of the context and the given timeout.
// The context given to the operation will have a timeout of a percentage of the overall timeout.
// The percentage is calculated from the given minimum number of attempts.
// If the given minimum number of attempts is 3, the timeout of each `op` call if the overall timeout / 3.
// The default minimum number of attempts is 2.
func RetryWithContext(ctx context.Context, op func(ctx context.Context) error, timeout time.Duration, minAttempts ...int) error {
	deadline, ok := ctx.Deadline()
	if ok {
		ctxTimeout := time.Until(deadline)
		if ctxTimeout < timeout {
			timeout = ctxTimeout
		}
	}
	divider := 2
	if len(minAttempts) == 1 {
		divider = minAttempts[0]
	}
	ctxOp := func() error {
		lctx, cancel := context.WithTimeout(ctx, timeout/time.Duration(divider))
		defer cancel()
		if err := op(lctx); err != nil {
			return errors.WithStack(err)
		}
		return nil
	}
	if err := retry(ctx, ctxOp, timeout); err != nil {
		return errors.WithStack(err)
	}
	return nil
}
