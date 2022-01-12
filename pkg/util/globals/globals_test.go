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

package globals

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Globals(t *testing.T) {
	t.Run("Defaults", func(t *testing.T) {
		require.EqualValues(t, DefaultKubernetesRequestBatchSize, GetGlobals().Kubernetes().RequestBatchSize().Get())
		require.EqualValues(t, DefaultKubernetesTimeout, GetGlobals().Timeouts().Kubernetes().Get())
		require.EqualValues(t, DefaultArangoDTimeout, GetGlobals().Timeouts().ArangoD().Get())
		require.EqualValues(t, DefaultReconciliationTimeout, GetGlobals().Timeouts().Reconciliation().Get())
		require.EqualValues(t, DefaultBackupConcurrentUploads, GetGlobals().Backup().ConcurrentUploads().Get())
	})

	t.Run("Override", func(t *testing.T) {
		GetGlobals().Kubernetes().RequestBatchSize().Set(0)
		GetGlobals().Timeouts().Kubernetes().Set(0)
		GetGlobals().Timeouts().ArangoD().Set(0)
		GetGlobals().Timeouts().Reconciliation().Set(0)
		GetGlobals().Backup().ConcurrentUploads().Set(0)
	})

	t.Run("Check", func(t *testing.T) {
		require.EqualValues(t, 0, GetGlobals().Kubernetes().RequestBatchSize().Get())
		require.EqualValues(t, 0, GetGlobals().Timeouts().Kubernetes().Get())
		require.EqualValues(t, 0, GetGlobals().Timeouts().ArangoD().Get())
		require.EqualValues(t, 0, GetGlobals().Timeouts().Reconciliation().Get())
		require.EqualValues(t, 0, GetGlobals().Backup().ConcurrentUploads().Get())
	})
}
