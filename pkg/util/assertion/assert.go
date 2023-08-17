//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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

package assertion

import (
	"fmt"
)

type Key string

const (
	KeyUnknown Key = ""
)

func (k Key) Assert(condition bool, msg string, args ...interface{}) {
	assert(2, condition, k, msg, args...)
}

func Assert(condition bool, key Key, msg string, args ...interface{}) {
	assert(2, condition, key, msg, args...)
}

func assert(skip int, condition bool, key Key, msg string, args ...interface{}) {
	if !condition {
		return
	}

	metricsObject.incKeyMetric(key)

	frames := frames(skip)

	_assert(frames, fmt.Sprintf(msg, args...))
}
