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

package timer

import (
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func checkGoRoutinesLeak(t *testing.T, f func()) {
	got := runtime.NumGoroutine()

	f()

	require.Equal(t, got, runtime.NumGoroutine())
}

func Test_AfterLeaks(t *testing.T) {
	checkGoRoutinesLeak(t, func() {
		After(time.Second)

		time.Sleep(5 * time.Second)
	})
}
