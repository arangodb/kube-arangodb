//
// DISCLAIMER
//
// Copyright 2025 ArangoDB GmbH, Cologne, Germany
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
	"testing"

	"github.com/stretchr/testify/require"
)

type testCauseError string

func (t testCauseError) Error() string {
	return string(t)

}

func Test_Causer(t *testing.T) {
	var err error = testCauseError("tete")

	v, ok := ExtractCause[testCauseError](err)
	require.True(t, ok)
	require.EqualValues(t, "tete", v)

	err = WithMessage(err, "msg")

	v, ok = ExtractCause[testCauseError](err)
	require.True(t, ok)
	require.EqualValues(t, "tete", v)

	v, ok = ExtractCause[testCauseError](Errorf("TEST"))
	require.False(t, ok)
	require.EqualValues(t, "", v)
}
