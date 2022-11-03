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

package tests

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func DurationBetween() func(t *testing.T, expected time.Duration, skew float64) {
	start := time.Now()
	return func(t *testing.T, expected time.Duration, skew float64) {
		current := time.Since(start)
		min := time.Duration(float64(expected) * (1 - skew))
		max := time.Duration(float64(expected) * (1 + skew))

		if current > max || current < min {
			require.Failf(t, "Skew is too big", "Expected %d, got %d", expected, current)
		}
	}
}
