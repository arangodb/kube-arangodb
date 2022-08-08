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

package timezones

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_Timezone(t *testing.T) {
	for tz, tq := range timezones {
		t.Run(tz, func(t *testing.T) {
			t.Run("Check fields", func(t *testing.T) {
				d, ok := GetTimezone(tz)
				require.True(t, ok)
				require.NotEmpty(t, d.Name)
				require.NotEmpty(t, d.Zone)
				require.NotEmpty(t, d.Parent)
			})

			t.Run("Check data", func(t *testing.T) {
				_, ok := timezonesData[tq.Parent]
				require.True(t, ok)
			})

			t.Run("Ensure Timezone will be loaded", func(t *testing.T) {
				tz, ok := tq.GetData()
				require.True(t, ok)
				l, err := time.LoadLocationFromTZData("", tz)
				require.NoError(t, err)

				z, offset := time.Now().In(l).Zone()
				require.Equal(t, tq.Zone, z)

				require.Equal(t, int(tq.Offset/time.Second), offset)
			})
		})
	}
}
