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

// AggregateError is a helper to wrap zero or more errors as a go `error`.
type AggregateError struct {
	errors []error
}

// Add returns a new error with given error added.
func (ae AggregateError) Add(e error) AggregateError {
	return AggregateError{
		errors: append(ae.errors, e),
	}
}

func (ae AggregateError) Error() string {
	switch len(ae.errors) {
	case 0:
		return "no errors"
	case 1:
		return ae.errors[0].Error()
	default:
		return ae.errors[0].Error() + ", ..."
	}
}

// AsError returns the given aggregate error if it contains 2 or more errors.
// It returns the first error if it contains exactly 1 error.
// Otherwise nil is returned.
func (ae AggregateError) AsError() error {
	if len(ae.errors) == 0 {
		return nil
	}
	if len(ae.errors) == 1 {
		return ae.errors[0]
	}
	return ae
}
