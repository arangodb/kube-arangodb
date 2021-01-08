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

package errors

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"syscall"

	errs "github.com/pkg/errors"

	"github.com/rs/zerolog"

	driver "github.com/arangodb/go-driver"
)

var (
	Cause     = errs.Cause
	New       = errs.New
	WithStack = errs.WithStack
	Wrap      = errs.Wrap
	Wrapf     = errs.Wrapf
)

func Newf(format string, args ...interface{}) error {
	return New(fmt.Sprintf(format, args...))
}

// WithMessage annotates err with a new message.
// The messages of given error is hidden.
// If err is nil, WithMessage returns nil.
func WithMessage(err error, message string) error {
	if err == nil {
		return nil
	}
	return &withMessage{
		cause: err,
		msg:   message,
	}
}

type withMessage struct {
	cause error
	msg   string
}

func (w *withMessage) Error() string { return w.msg }
func (w *withMessage) Cause() error  { return w.cause }

func (w *withMessage) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprintf(s, "%+v\n", w.Cause())
			io.WriteString(s, w.msg)
			return
		}
		fallthrough
	case 's', 'q':
		io.WriteString(s, w.Error())
	}
}

type timeout interface {
	Timeout() bool
}

// IsTimeout returns true if the given error is caused by a timeout error.
func IsTimeout(err error) bool {
	if err == nil {
		return false
	}
	if t, ok := errs.Cause(err).(timeout); ok {
		return t.Timeout()
	}
	return false
}

type temporary interface {
	Temporary() bool
}

// IsTemporary returns true if the given error is caused by a temporary error.
func IsTemporary(err error) bool {
	if err == nil {
		return false
	}
	if t, ok := errs.Cause(err).(temporary); ok {
		return t.Temporary()
	}
	return false
}

// IsEOF returns true if the given error is caused by an EOF error.
func IsEOF(err error) bool {
	err = errs.Cause(err)
	if err == io.EOF {
		return true
	}
	if ok, err := libCause(err); ok {
		return IsEOF(err)
	}
	return false
}

// IsConnectionRefused returns true if the given error is caused by an "connection refused" error.
func IsConnectionRefused(err error) bool {
	err = errs.Cause(err)
	if err, ok := err.(syscall.Errno); ok {
		return err == syscall.ECONNREFUSED
	}
	if ok, err := libCause(err); ok {
		return IsConnectionRefused(err)
	}
	return false
}

// IsConnectionReset returns true if the given error is caused by an "connection reset by peer" error.
func IsConnectionReset(err error) bool {
	err = errs.Cause(err)
	if err, ok := err.(syscall.Errno); ok {
		return err == syscall.ECONNRESET
	}
	if ok, err := libCause(err); ok {
		return IsConnectionReset(err)
	}
	return false
}

// IsContextCanceled returns true if the given error is caused by a context cancelation.
func IsContextCanceled(err error) bool {
	err = errs.Cause(err)
	if err == context.Canceled {
		return true
	}
	if ok, err := libCause(err); ok {
		return IsContextCanceled(err)
	}
	return false
}

// IsContextDeadlineExpired returns true if the given error is caused by a context deadline expiration.
func IsContextDeadlineExpired(err error) bool {
	err = errs.Cause(err)
	if err == context.DeadlineExceeded {
		return true
	}
	if ok, err := libCause(err); ok {
		return IsContextDeadlineExpired(err)
	}
	return false
}

// IsContextCanceledOrExpired returns true if the given error is caused by a context cancelation
// or deadline expiration.
func IsContextCanceledOrExpired(err error) bool {
	err = errs.Cause(err)
	if err == context.Canceled || err == context.DeadlineExceeded {
		return true
	}
	if ok, err := libCause(err); ok {
		return IsContextCanceledOrExpired(err)
	}
	return false
}

// libCause returns the Cause of well known go library errors.
func libCause(err error) (bool, error) {
	original := err
	for {
		switch e := err.(type) {
		case *driver.ResponseError:
			err = e.Err
		case *net.DNSConfigError:
			err = e.Err
		case *net.OpError:
			err = e.Err
		case *os.SyscallError:
			err = e.Err
		case *url.Error:
			err = e.Err
		default:
			return err != original, err
		}
	}
}

func LogError(logger zerolog.Logger, msg string, f func() error) {
	if err := f(); err != nil {
		logger.Error().Err(err).Msg(msg)
	}
}

type causer interface {
	Cause() error
}

func IsReconcile(err error) bool {
	if err == nil {
		return false
	}

	if _, ok := err.(reconcile); ok {
		return true
	}

	if c, ok := err.(causer); ok {
		return IsReconcile(c.Cause())
	}

	return false
}

func Reconcile() error {
	return reconcile{}
}

type reconcile struct {
}

func (r reconcile) Error() string {
	return "reconcile"
}
