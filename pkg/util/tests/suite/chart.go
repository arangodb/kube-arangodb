//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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

package suite

import (
	_ "embed"
	"testing"

	"github.com/stretchr/testify/require"
)

//go:embed chart/example-1.0.0.tgz
var chart_example_1_0_0 []byte

//go:embed chart/example-1.0.1.tgz
var chart_example_1_0_1 []byte

//go:embed chart/example-1.1.0.tgz
var chart_example_1_1_0 []byte

func GetChart(t *testing.T, name, version string) []byte {
	switch name {
	case "example":
		switch version {
		case "1.0.0":
			return chart_example_1_0_0
		case "1.0.1":
			return chart_example_1_0_1
		case "1.1.0":
			return chart_example_1_1_0
		}
	}

	require.Fail(t, "Chart with version not found")

	return nil
}
