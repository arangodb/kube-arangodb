//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
)

func TestIsActionTimeout(t *testing.T) {
	type testCase struct {
		timeout        api.Timeout
		action         api.Action
		expectedResult bool
	}

	timeFiveMinutesAgo := meta.Time{
		Time: time.Now().Add(-time.Hour),
	}

	testCases := map[string]testCase{
		"nil start time": {
			timeout:        api.Timeout{},
			action:         api.Action{},
			expectedResult: false,
		},
		"infinite timeout": {
			timeout:        api.NewTimeout(0),
			action:         api.Action{},
			expectedResult: false,
		},
		"timeouted case": {
			timeout: api.NewTimeout(time.Minute),
			action: api.Action{
				StartTime: &timeFiveMinutesAgo,
			},
			expectedResult: true,
		},
		"still in progress case": {
			timeout: api.NewTimeout(time.Minute * 10),
			action: api.Action{
				StartTime: &timeFiveMinutesAgo,
			},
			expectedResult: true,
		},
	}

	for n, c := range testCases {
		t.Run(n, func(t *testing.T) {
			require.Equal(t, c.expectedResult, isActionTimeout(c.timeout, c.action))
		})
	}
}
