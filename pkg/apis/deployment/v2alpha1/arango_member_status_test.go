//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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

package v2alpha1

import (
	"testing"

	"github.com/stretchr/testify/require"
	core "k8s.io/api/core/v1"
)

func Test_ArangoMemberStatus_Propagate(t *testing.T) {
	var status ArangoMemberStatus

	t.Run("Add first condition", func(t *testing.T) {
		var member = MemberStatus{
			Conditions: ConditionList{
				{
					Type:   ConditionTypeScheduled,
					Status: core.ConditionTrue,
				},
			},
		}

		require.False(t, status.InSync(member))
		require.True(t, status.Propagate(member))
		require.True(t, status.InSync(member))
		require.False(t, status.Propagate(member))

		require.Equal(t, member.Conditions, status.Conditions)
	})

	t.Run("Update first condition", func(t *testing.T) {
		var member = MemberStatus{
			Conditions: ConditionList{
				{
					Type:   ConditionTypeScheduled,
					Status: core.ConditionFalse,
				},
			},
		}

		require.NotEqual(t, member.Conditions, status.Conditions)

		require.False(t, status.InSync(member))
		require.True(t, status.Propagate(member))
		require.True(t, status.InSync(member))
		require.False(t, status.Propagate(member))

		require.Equal(t, member.Conditions, status.Conditions)
	})
}
