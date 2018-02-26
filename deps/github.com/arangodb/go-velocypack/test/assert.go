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
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func ASSERT_EQ(a, b interface{}, t *testing.T) {
	as, asOk := a.(string)
	bs, bsOk := b.(string)
	if asOk && bsOk {
		if strings.Compare(as, bs) != 0 {
			t.Errorf("Expected '%s', '%s' to be equal\nat %s", as, bs, callerInfo(2))
		}
	} else if !reflect.DeepEqual(a, b) {
		t.Errorf("Expected %v, %v to be equal\nat %s", a, b, callerInfo(2))
	}
}

func ASSERT_DOUBLE_EQ(a, b float64, t *testing.T) {
	if a != b {
		t.Errorf("Expected %v, %v to be equal\nat %s", a, b, callerInfo(2))
	}
}

func ASSERT_TRUE(a bool, t *testing.T) {
	if !a {
		t.Errorf("Expected true\nat %s", callerInfo(2))
	}
}

func ASSERT_FALSE(a bool, t *testing.T) {
	if a {
		t.Errorf("Expected false\nat %s", callerInfo(2))
	}
}

func ASSERT_NIL(a interface{}, t *testing.T) {
	if a != nil {
		t.Errorf("Expected nil, got %v\nat %s", a, callerInfo(2))
	}
}

func ASSERT_VELOCYPACK_EXCEPTION(errorType func(error) bool, t *testing.T) func(args ...interface{}) {
	return func(args ...interface{}) {
		l := len(args)
		if l == 0 {
			t.Fatalf("Expected at least 1 error argument\nat %s", callerInfo(2))
		}
		last := args[l-1]
		if last == nil {
			t.Errorf("Expected error, got nil\nat %s", callerInfo(2))
		} else if err, ok := last.(error); !ok {
			t.Fatalf("Expected last argument to be of type error, got %v\nat %s", args[l-1], callerInfo(2))
		} else if !errorType(err) {
			t.Errorf("Expected error, got %s\nat %s", err, callerInfo(2))
		}
	}
}

func callerInfo(depth int) string {
	_, file, line, ok := runtime.Caller(depth)
	if !ok {
		return "?"
	}
	return fmt.Sprintf("%s (%d)", file, line)
}
