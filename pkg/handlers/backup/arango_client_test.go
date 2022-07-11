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

package backup

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/arangodb/go-driver"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func Test_Errors_Temporary(t *testing.T) {
	// Arrange
	type check struct {
		err       error
		temporary bool
	}

	checks := map[string]check{}

	for _, code := range temporaryErrorNum {
		checks[fmt.Sprintf("errorNum-%d", code)] = check{
			err: driver.ArangoError{
				ErrorNum: code,
			},
			temporary: true,
		}

		checks[fmt.Sprintf("errorNumCaused-%d", code)] = check{
			err: errors.WithStack(driver.ArangoError{
				ErrorNum: code,
			}),
			temporary: true,
		}
	}

	// generate some other data
	for id := 0; id < 8; {
		n := rand.Intn(30000)

		if temporaryErrorNum.Has(n) {
			continue
		}

		name := fmt.Sprintf("nonExpectedErrorNum-%d", n)

		if _, ok := checks[name]; ok {
			continue
		}

		id++

		checks[name] = check{
			err: driver.ArangoError{
				ErrorNum: n,
			},
			temporary: false,
		}
	}

	for _, code := range temporaryCodes {
		checks[fmt.Sprintf("code-%d", code)] = check{
			err: errors.WithStack(driver.ArangoError{
				Code: code,
			}),
			temporary: true,
		}

		checks[fmt.Sprintf("codeCaused-%d", code)] = check{
			err: driver.ArangoError{
				Code: code,
			},
			temporary: true,
		}
	}

	// Act
	for testName, c := range checks {
		t.Run(testName, func(t *testing.T) {
			res := isTemporaryError(c.err)
			if c.temporary {
				require.True(t, res)
			} else {
				require.False(t, res)
			}
		})
	}
}
