//
// Copyright 2017 ArangoDB GmbH, Cologne, Germany
//
// The Programs (which include both the software and documentation) contain
// proprietary information of ArangoDB GmbH; they are provided under a license
// agreement containing restrictions on use and disclosure and are also
// protected by copyright, patent and other intellectual and industrial
// property laws. Reverse engineering, disassembly or decompilation of the
// Programs, except to the extent required to obtain interoperability with
// other independently created software or as specified by law, is prohibited.
//
// It shall be the licensee's responsibility to take all appropriate fail-safe,
// backup, redundancy, and other measures to ensure the safe use of
// applications if the Programs are used for purposes such as nuclear,
// aviation, mass transit, medical, or other inherently dangerous applications,
// and ArangoDB GmbH disclaims liability for any damages caused by such use of
// the Programs.
//
// This software is the confidential and proprietary information of ArangoDB
// GmbH. You shall not disclose such confidential and proprietary information
// and shall use it only in accordance with the terms of the license agreement
// you entered into with ArangoDB GmbH.
//
// Author Ewout Prangsma
//

package retry

import (
	"context"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/pkg/errors"
)

var (
	maskAny = errors.WithStack
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
	type causer interface {
		Cause() error
	}

	for err != nil {
		if pe, ok := err.(*permanentError); ok {
			return pe, true
		}
		cause, ok := err.(causer)
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
		return maskAny(err)
	}
	if failure != nil {
		return maskAny(failure)
	}
	return nil
}

// Retry the given operation until it succeeds,
// has a permanent failure or times out.
func Retry(op func() error, timeout time.Duration) error {
	return retry(nil, op, timeout)
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
			return maskAny(err)
		}
		return nil
	}
	if err := retry(ctx, ctxOp, timeout); err != nil {
		return maskAny(err)
	}
	return nil
}
