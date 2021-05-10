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
// Author Adam Janikowski
//

package shared

import (
	"fmt"
	"io"
	"strings"
)

type ResourceError struct {
	Prefix string
	Err    error
}

// Error return string representation of error
func (p ResourceError) Error() string {
	return fmt.Sprintf("%s: %s", p.Prefix, p.Err.Error())
}

// Format formats error with verbs
func (p *ResourceError) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprintf(s, "%s\n", p.Error())
			fmt.Fprintf(s, "%+v", p.Err)
			return
		}
		fallthrough
	case 's':
		io.WriteString(s, p.Error())
	case 'q':
		fmt.Fprintf(s, "%q", p.Error())
	}
}

// PrefixResourceError creates new prefixed error. If error is already prefixed then current key is appended
func PrefixResourceError(prefix string, err error) error {
	if err == nil {
		return nil
	}

	switch e := err.(type) {
	case ResourceError:
		return ResourceError{
			Prefix: fmt.Sprintf("%s.%s", prefix, e.Prefix),
			Err:    e.Err,
		}
	default:
		return ResourceError{
			Prefix: prefix,
			Err:    err,
		}
	}
}

// PrefixResourceErrors creates new prefixed errors. If error is already prefixed then current key is appended
func PrefixResourceErrors(prefix string, errs ...error) error {
	prefixed := make([]error, 0, len(errs))

	for _, err := range errs {
		switch errType := err.(type) {
		case MergedErrors:
			for _, subError := range errType.errors {
				prefixed = append(prefixed, PrefixResourceError(prefix, subError))
			}
		default:
			prefixed = append(prefixed, PrefixResourceError(prefix, err))
		}
	}

	return WithErrors(prefixed...)
}

type MergedErrors struct {
	errors []error
}

func (m MergedErrors) Error() string {
	errStrings := make([]string, 0, len(m.errors))

	for _, err := range m.errors {
		if err == nil {
			continue
		}

		errStrings = append(errStrings, err.Error())
	}

	return fmt.Sprintf("Received %d errors: %s", len(errStrings), strings.Join(errStrings, ", "))
}

func (m MergedErrors) Errors() []error {
	return m.errors
}

// WithErrors filter out nil errors
func WithErrors(errs ...error) error {
	filteredErrs := make([]error, 0, len(errs))

	for _, err := range errs {
		if err == nil {
			continue
		}

		switch errType := err.(type) {
		case MergedErrors:
			filteredErrs = append(filteredErrs, errType.errors...)
		default:
			filteredErrs = append(filteredErrs, err)
		}
	}

	if len(filteredErrs) == 0 {
		return nil
	}

	return MergedErrors{
		errors: filteredErrs,
	}
}
