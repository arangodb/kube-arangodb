//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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
//

package provisioner

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

var (
	// BadRequestError indicates invalid arguments.
	BadRequestError = StatusError{StatusCode: http.StatusBadRequest, message: "bad request"}
	// InternalServerError indicates an unspecified error inside the server, perhaps a bug.
	InternalServerError = StatusError{StatusCode: http.StatusInternalServerError, message: "internal server error"}
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

// IsBadRequest returns true if the given error is caused by a BadRequestError.
func IsBadRequest(err error) bool {
	return IsStatusErrorWithCode(err, http.StatusBadRequest)
}

// IsInternalServer returns true if the given error is caused by a InternalServerError.
func IsInternalServer(err error) bool {
	return IsStatusErrorWithCode(err, http.StatusInternalServerError)
}

// ParseResponseError returns an error from given response.
// It tries to parse the body (if given body is nil, will be read from response)
// for ErrorResponse.
func ParseResponseError(r *http.Response, body []byte) error {
	// Read body (if needed)
	if body == nil {
		defer r.Body.Close()
		body, _ = io.ReadAll(r.Body)
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

	return result
}
