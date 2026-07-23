//
// DISCLAIMER
//
// Copyright 2026 ArangoDB GmbH, Cologne, Germany
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

package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_ArangoPermissionBindingRef(t *testing.T) {
	t.Run("Validate", func(t *testing.T) {
		require.Error(t, (*ArangoPermissionBindingRef)(nil).Validate())
		require.Error(t, (&ArangoPermissionBindingRef{}).Validate(), "neither name nor direct")
		require.Error(t, (&ArangoPermissionBindingRef{Name: "r", Direct: "managed:predefined:coredb-reader"}).Validate(), "both set is mutually exclusive")
		require.NoError(t, (&ArangoPermissionBindingRef{Name: "r"}).Validate())
		require.NoError(t, (&ArangoPermissionBindingRef{Direct: "managed:predefined:coredb-reader"}).Validate())
	})

	t.Run("IsDirect", func(t *testing.T) {
		require.False(t, (*ArangoPermissionBindingRef)(nil).IsDirect())
		require.False(t, (&ArangoPermissionBindingRef{Name: "r"}).IsDirect())
		require.True(t, (&ArangoPermissionBindingRef{Direct: "managed:predefined:coredb-reader"}).IsDirect())
	})

	t.Run("GetReference", func(t *testing.T) {
		require.Equal(t, "", (*ArangoPermissionBindingRef)(nil).GetReference())
		require.Equal(t, "r", (&ArangoPermissionBindingRef{Name: "r"}).GetReference())
		// Direct is used as-is and takes precedence over the resolved CRD name.
		require.Equal(t, "managed:predefined:coredb-reader", (&ArangoPermissionBindingRef{Direct: "managed:predefined:coredb-reader"}).GetReference())
	})

	t.Run("Hash reacts to direct", func(t *testing.T) {
		require.NotEqual(t,
			(&ArangoPermissionBindingRef{Name: "coredb-reader"}).Hash(),
			(&ArangoPermissionBindingRef{Direct: "coredb-reader"}).Hash(),
			"a CRD name and a direct name must not collide")
	})
}
