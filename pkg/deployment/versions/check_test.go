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

package versions

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/arangodb/go-driver"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
)

func runCheckTest(t *testing.T, name string, version driver.Version, enterprise bool, expected bool, checker ...func(c Check) Check) {
	t.Run(name, func(t *testing.T) {
		c := NewCheck(api.ImageInfo{
			ArangoDBVersion: version,
			Enterprise:      enterprise,
		})

		for _, f := range checker {
			c = f(c)
		}

		if expected {
			require.True(t, c.Evaluate())
		} else {
			require.False(t, c.Evaluate())
		}
	})
}

func Test_Check(t *testing.T) {
	runCheckTest(t, "Empty", "3.6.0", true, true)

	runCheckTest(t, "Enterprise - ok", "3.6.0", true, true, func(c Check) Check {
		return c.Enterprise()
	})
	runCheckTest(t, "Enterprise - invalid", "3.6.0", false, false, func(c Check) Check {
		return c.Enterprise()
	})

	runCheckTest(t, "Community - ok", "3.6.0", false, true, func(c Check) Check {
		return c.Community()
	})
	runCheckTest(t, "Community - invalid", "3.6.0", true, false, func(c Check) Check {
		return c.Community()
	})

	runCheckTest(t, "3.6.5 <=", "3.6.5", true, true, func(c Check) Check {
		return c.AboveOrEqual("3.6.5")
	})

	runCheckTest(t, "3.6.5 <", "3.6.5", true, false, func(c Check) Check {
		return c.Above("3.6.5")
	})

	runCheckTest(t, "3.6.5 >=", "3.6.5", true, true, func(c Check) Check {
		return c.AboveOrEqual("3.6.5")
	})

	runCheckTest(t, "3.6.5 >", "3.6.5", true, false, func(c Check) Check {
		return c.Above("3.6.5")
	})
}
