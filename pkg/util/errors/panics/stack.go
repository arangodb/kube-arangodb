//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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

package panics

import (
	"fmt"
	"runtime"
)

func GetStack(skip int) StackEntries {
	frames := make([]uintptr, 30)

	f := runtime.Callers(skip, frames)
	fr := runtime.CallersFrames(frames[0:f])
	ret := make(StackEntries, f)

	id := 0
	for {
		f, ok := fr.Next()
		if !ok {
			ret = ret[0:id]
			break
		}

		ret[id] = StackEntry{
			File:     f.File,
			Function: f.Function,
			Line:     f.Line,
		}

		id++
	}

	return ret
}

type StackEntries []StackEntry

func (s StackEntries) String() []string {
	r := make([]string, len(s))

	for id := range s {
		r[id] = s[id].String()
	}

	return r
}

type StackEntry struct {
	File, Function string
	Line           int
}

func (s StackEntry) String() string {
	return fmt.Sprintf("%s:%d - %s()", s.File, s.Line, s.Function)
}
