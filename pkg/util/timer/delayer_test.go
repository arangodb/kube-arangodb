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
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func withTime(f func()) time.Duration {
	now := time.Now()
	f()
	return time.Since(now)
}

func Test_Delayer(t *testing.T) {
	d := NewDelayer()

	t.Run("Ensure instant execution", func(t *testing.T) {
		require.True(t, withTime(func() {
			d.Wait()
		}) < time.Millisecond)

		require.True(t, withTime(func() {
			d.Wait()
		}) < time.Millisecond)
	})

	t.Run("Delay execution", func(t *testing.T) {
		require.True(t, withTime(func() {
			d.Delay(50 * time.Millisecond)
			d.Wait()
		}) >= 50*time.Millisecond)
	})

	t.Run("Delay execution, but allow multiple ones", func(t *testing.T) {
		require.True(t, withTime(func() {
			d.Delay(50 * time.Millisecond)
			d.Wait()
			d.Wait()
			d.Wait()
			d.Wait()
		}) >= 50*time.Millisecond)
	})

	t.Run("Delay execution multiple times", func(t *testing.T) {
		require.True(t, withTime(func() {
			d.Delay(50 * time.Millisecond)
			d.Wait()
			d.Delay(50 * time.Millisecond)
			d.Wait()
			d.Delay(50 * time.Millisecond)
			d.Wait()
			d.Delay(50 * time.Millisecond)
			d.Wait()
		}) >= 200*time.Millisecond)
	})
}
