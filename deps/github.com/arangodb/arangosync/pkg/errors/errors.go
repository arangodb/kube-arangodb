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

package errors

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"syscall"

	driver "github.com/arangodb/go-driver"
	errs "github.com/pkg/errors"
)

var (
	Cause     = errs.Cause
	New       = errs.New
	WithStack = errs.WithStack
	Wrap      = errs.Wrap
	Wrapf     = errs.Wrapf
)

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
