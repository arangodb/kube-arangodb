//
// DISCLAIMER
//
// Copyright 2016-2025 ArangoDB GmbH, Cologne, Germany
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

package errors

import (
	"errors"
	"fmt"
	"io"
	goStrings "strings"
)

func ExpandArray(err error) []error {
	if err == nil {
		return nil
	}

	var v Array
	if errors.As(err, &v) {
		return v
	}

	return []error{err}
}

type Array []error

func (a Array) Error() string {
	q := make([]string, len(a))

	for id := range a {
		q[id] = a[id].Error()
	}

	return fmt.Sprintf("Received %d errors: %s", len(q), goStrings.Join(q, ", "))
}

// Format formats error with verbs
func (p Array) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprintf(s, "%s\n", p.Error())
			for _, err := range p {
				fmt.Fprintf(s, "%+v\n", err)
			}
			return
		}
		fallthrough
	case 's':
		io.WriteString(s, p.Error())
	case 'q':
		fmt.Fprintf(s, "%q", p.Error())
	}
}

func Errors(errs ...error) error {
	f := make(Array, 0, len(errs))

	for _, err := range errs {
		if err == nil {
			continue
		}

		f = append(f, ExpandArray(err)...)
	}

	if len(f) == 0 {
		return nil
	}

	return f
}
