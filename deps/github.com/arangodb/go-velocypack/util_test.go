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

package velocypack

import "testing"

func TestAlignAt(t *testing.T) {
	tests := []struct {
		Value     uint
		Alignment uint
		Expected  uint
	}{
		{10, 16, 16},
		{16, 16, 16},
		{2345, 4096, 4096},
		{7000, 4096, 4096 * 2},
	}
	for _, test := range tests {
		result := alignAt(test.Value, test.Alignment)
		if result != test.Expected {
			t.Errorf("alignAt(%d, %d) failed. Expected %d, got %d", test.Value, test.Alignment, test.Expected, result)
		}
	}
}
