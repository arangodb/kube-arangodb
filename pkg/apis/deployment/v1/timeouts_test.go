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

package v1

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"
)

func Test_API_Timeouts(t *testing.T) {
	d := `
spec:
  timeouts:
    actions:
      CleanOutMember: 50h
`

	var q ArangoDeployment

	require.NoError(t, yaml.Unmarshal([]byte(d), &q))

	require.EqualValues(t, q.Spec.Timeouts.Actions[ActionTypeCleanOutMember].Duration, time.Hour*50)

	z, err := yaml.Marshal(q)
	require.NoError(t, err)

	require.NoError(t, yaml.Unmarshal(z, &q))

	require.EqualValues(t, q.Spec.Timeouts.Actions[ActionTypeCleanOutMember].Duration, time.Hour*50)
}
