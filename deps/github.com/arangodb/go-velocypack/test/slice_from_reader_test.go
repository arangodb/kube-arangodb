//
// DISCLAIMER
//
// Copyright 2017 ArangoDB GmbH, Cologne, Germany
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

package test

import (
	"bufio"
	"bytes"
	"fmt"
	"testing"

	velocypack "github.com/arangodb/go-velocypack"
)

// assertEqualFromReader wraps the given slice in a byte Buffer (the io.Reader) and
// calls SliceFromReader on that.
// It then compares the 2 slices.
func assertEqualFromReader(t *testing.T, s velocypack.Slice, args ...interface{}) {
	// Normal reader
	{
		buf := bytes.NewBuffer(s)
		s2, err := velocypack.SliceFromReader(buf)
		var msg string
		if len(args) > 0 {
			msg = " (" + fmt.Sprintf(args[0].(string), args[1:]...) + ")"
		}
		if err != nil {
			t.Errorf("SliceFromReader failed at %s: %v%s", callerInfo(2), err, msg)
		} else if s.String() != s2.String() {
			t.Errorf("SliceFromReader return different slice at %s. Got:\n\t'%s', expected:\n\t'%s'%s", callerInfo(2), s2.String(), s.String(), msg)
		}
	}

	// Buffered reader
	{
		brd := bufio.NewReader(bytes.NewBuffer(s))
		s2, err := velocypack.SliceFromReader(brd)
		var msg string
		if len(args) > 0 {
			msg = " (" + fmt.Sprintf(args[0].(string), args[1:]...) + ")"
		}
		if err != nil {
			t.Errorf("SliceFromReader failed at %s: %v%s", callerInfo(2), err, msg)
		} else if s.String() != s2.String() {
			t.Errorf("SliceFromReader return different slice at %s. Got:\n\t'%s', expected:\n\t'%s'%s", callerInfo(2), s2.String(), s.String(), msg)
		}
	}
}
