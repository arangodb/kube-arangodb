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

package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"

	"github.com/arangodb/arangosync/pkg/retry"
)

var (
	maskAny = errors.WithStack
	// NotFoundError indicates that an object does not exist.
	NotFoundError = StatusError{StatusCode: http.StatusNotFound, message: "not found"}
	// ServiceUnavailableError indicates that right now the service is not available, please retry later.
	ServiceUnavailableError = StatusError{StatusCode: http.StatusServiceUnavailable, message: "service unavailable"}
	// BadRequestError indicates invalid arguments.
	BadRequestError = StatusError{StatusCode: http.StatusBadRequest, message: "bad request"}
	// PreconditionFailedError indicates that the state of the system is such that the request cannot be executed.
	PreconditionFailedError = StatusError{StatusCode: http.StatusPreconditionFailed, message: "precondition failed"}
	// InternalServerError indicates an unspecified error inside the server, perhaps a bug.
	InternalServerError = StatusError{StatusCode: http.StatusInternalServerError, message: "internal server error"}
	// UnauthorizedError indicates that the request has not the correct authorization.
	UnauthorizedError = StatusError{StatusCode: http.StatusUnauthorized, message: "unauthorized"}
	// RequestTimeoutError indicates that the request is taken longer than we're prepared to wait.
	RequestTimeoutError = StatusError{StatusCode: http.StatusRequestTimeout, message: "request timeout"}
)

type StatusError struct {
	StatusCode int
	message    string
}

func (e StatusError) Error() string {
	if e.message != "" {
		return e.message
	}
	return fmt.Sprintf("Status %d", e.StatusCode)
}

// IsStatusError returns the status code and true
// if the given error is caused by a StatusError.
func IsStatusError(err error) (int, bool) {
	err = errors.Cause(err)
	if serr, ok := err.(StatusError); ok {
		return serr.StatusCode, true
	}
	return 0, false
}

// IsStatusErrorWithCode returns true if the given error is caused
// by a StatusError with given code.
func IsStatusErrorWithCode(err error, code int) bool {
	err = errors.Cause(err)
	if serr, ok := err.(StatusError); ok {
		return serr.StatusCode == code
	}
	return false
}

type ErrorResponse struct {
	Error string
}

type RedirectToError struct {
	Location string
}

func (e RedirectToError) Error() string {
	return fmt.Sprintf("Redirect to: %s", e.Location)
}

// IsRedirectTo returns true when the given error is caused by an
// RedirectToError. If so, it also returns the redirect location.
func IsRedirectTo(err error) (string, bool) {
	err = errors.Cause(err)
	if rterr, ok := err.(RedirectToError); ok {
		return rterr.Location, true
	}
	return "", false
}

// IsNotFound returns true if the given error is caused by a NotFoundError.
func IsNotFound(err error) bool {
	return IsStatusErrorWithCode(err, http.StatusNotFound)
}

// IsServiceUnavailable returns true if the given error is caused by a ServiceUnavailableError.
func IsServiceUnavailable(err error) bool {
	return IsStatusErrorWithCode(err, http.StatusServiceUnavailable)
}

// IsBadRequest returns true if the given error is caused by a BadRequestError.
func IsBadRequest(err error) bool {
	return IsStatusErrorWithCode(err, http.StatusBadRequest)
}

// IsPreconditionFailed returns true if the given error is caused by a PreconditionFailedError.
func IsPreconditionFailed(err error) bool {
	return IsStatusErrorWithCode(err, http.StatusPreconditionFailed)
}

// IsInternalServer returns true if the given error is caused by a InternalServerError.
func IsInternalServer(err error) bool {
	return IsStatusErrorWithCode(err, http.StatusInternalServerError)
}

// IsUnauthorized returns true if the given error is caused by a UnauthorizedError.
func IsUnauthorized(err error) bool {
	return IsStatusErrorWithCode(err, http.StatusUnauthorized)
}

// IsRequestTimeout returns true if the given error is caused by a RequestTimeoutError.
func IsRequestTimeout(err error) bool {
	return IsStatusErrorWithCode(err, http.StatusRequestTimeout)
}

// IsCanceled returns true if the given error is caused by a context.Canceled.
func IsCanceled(err error) bool {
	return errors.Cause(err) == context.Canceled
}

// ParseResponseError returns an error from given response.
// It tries to parse the body (if given body is nil, will be read from response)
// for ErrorResponse.
func ParseResponseError(r *http.Response, body []byte) error {
	// Read body (if needed)
	if body == nil {
		defer r.Body.Close()
		body, _ = ioutil.ReadAll(r.Body)
	}
	return parseResponseError(body, r.StatusCode)
}

// parseResponseError returns an error from given response.
// It tries to parse the body (if given body is nil, will be read from response)
// for ErrorResponse.
func parseResponseError(body []byte, statusCode int) error {
	// Parse body (if available)
	var result error
	if len(body) > 0 {
		var errRes ErrorResponse
		if err := json.Unmarshal(body, &errRes); err == nil {
			// Found ErrorResponse
			result = StatusError{StatusCode: statusCode, message: errRes.Error}
		}
	}

	if result == nil {
		// No ErrorResponse found, fallback to default message
		result = StatusError{StatusCode: statusCode}
	}

	// Is permanent error?
	if statusCode >= 400 && statusCode < 500 {
		result = retry.Permanent(result)
	}

	return result
}
