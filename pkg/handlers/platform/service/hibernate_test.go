//
// DISCLAIMER
//
// Copyright 2025-2026 ArangoDB GmbH, Cologne, Germany
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

package service

import (
	"testing"

	"github.com/stretchr/testify/require"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	platformApi "github.com/arangodb/kube-arangodb/pkg/apis/platform/v1beta1"
)

func TestUpdateHibernatedCondition(t *testing.T) {
	var conditions api.ConditionList

	// Setting it the first time reports a change and reflects the requested state.
	require.True(t, updateHibernatedCondition(&conditions, true))
	require.True(t, conditions.IsTrue(platformApi.HibernatedCondition))

	// Setting the same value again is a no-op.
	require.False(t, updateHibernatedCondition(&conditions, true))

	// Clearing hibernation flips the condition and reports a change.
	require.True(t, updateHibernatedCondition(&conditions, false))
	require.False(t, conditions.IsTrue(platformApi.HibernatedCondition))

	c, ok := conditions.Get(platformApi.HibernatedCondition)
	require.True(t, ok, "Hibernated condition must be present even when false")
	require.False(t, updateHibernatedCondition(&conditions, false))
	_ = c
}
