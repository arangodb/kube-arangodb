//
// DISCLAIMER
//
// Copyright 2016-2026 ArangoDB GmbH, Cologne, Germany
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

package agency

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/arangodb/kube-arangodb/pkg/deployment/agency/poll"
	"github.com/arangodb/kube-arangodb/pkg/deployment/agency/state"
)

func Test_AgencyApply(t *testing.T) {
	s := poll.NewApplier[state.State](poll.ApplierConfig{AllowUnsupportedOperations: false})
	apply(t, "Ensure Missing Server RebootID", s, func(t *testing.T, obj *state.State) {
		id, ok := obj.GetRebootID("a")
		require.False(t, ok)
		require.EqualValues(t, 0, id)
	}, `{}`)
	apply(t, "Set Current Server RebootID", s, func(t *testing.T, obj *state.State) {
		id, ok := obj.GetRebootID("a")
		require.True(t, ok)
		require.EqualValues(t, 1, id)
	}, `{
  "Current/ServersKnown": {
     "op": "set",
     "new": {"a": {"rebootId": 1}}
   }
}`)
	apply(t, "Increment Server RebootID", s, func(t *testing.T, obj *state.State) {
		id, ok := obj.GetRebootID("a")
		require.True(t, ok)
		require.EqualValues(t, 2, id)
	}, `{
  "Current/ServersKnown/a/rebootId": {
     "op": "increment"
   }
}`)
}
func apply[T interface{}](t *testing.T, name string, obj poll.Applier[T], verify func(t *testing.T, obj *T), itemSet string) {
	t.Run(name, func(t *testing.T) {
		var i poll.ItemSet
		require.NoError(t, json.Unmarshal([]byte(itemSet), &i))
		require.NoError(t, obj.ApplyItemSet(i))
		if v := verify; v != nil {
			v(t, obj.Get())
		}
	})
}
