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

package v1

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

func TestArangoBackupStatusBackOff_Backoff(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		var spec *ArangoBackupSpecBackOff
		var status *ArangoBackupStatusBackOff

		n := status.Backoff(spec)

		require.Equal(t, 1, n.GetIterations())
		require.True(t, n.GetNext().After(time.Now().Add(time.Duration(9.9*float64(time.Second)))))
	})

	t.Run("Test MaxIterations", func(t *testing.T) {
		var spec = &ArangoBackupSpecBackOff{
			Iterations:    util.NewType[int](2),
			MaxIterations: util.NewType[int](3),
		}
		var status *ArangoBackupStatusBackOff

		n := status.Backoff(spec)
		require.Equal(t, 1, n.GetIterations())
		require.True(t, n.ShouldBackoff(spec))

		n.Iterations = 3
		n = n.Backoff(spec)
		require.Equal(t, 3, n.GetIterations())
		require.False(t, n.ShouldBackoff(spec))
	})
}
