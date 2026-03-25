//
// DISCLAIMER
//
// Copyright 2026 ArangoDB GmbH, Cologne, Germany
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

package reconcile

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_LicenseExpirationCalculation(t *testing.T) {
	check := func(name string, dur, expected time.Duration) {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, expected, calculateInternalLicenseExpiration(dur))
		})
	}

	check("Check Min Value", 30*time.Minute, 15*time.Minute)
	check("Check Max Value", 15*time.Minute, 10*time.Minute)
	check("Check Default Value", 60*time.Minute, 20*time.Minute)
}
