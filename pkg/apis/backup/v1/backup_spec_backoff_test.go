//
// DISCLAIMER
//
// Copyright 2016-2021 ArangoDB GmbH, Cologne, Germany
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

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/stretchr/testify/require"
)

func TestArangoBackupSpecBackOff_Backoff(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		var b *ArangoBackupSpecBackOff

		require.Equal(t, 10*time.Second, b.Backoff(0))
		require.Equal(t, 20*time.Second, b.Backoff(1))
		require.Equal(t, 30*time.Second, b.Backoff(2))
		require.Equal(t, 40*time.Second, b.Backoff(3))
		require.Equal(t, 50*time.Second, b.Backoff(4))
		require.Equal(t, 60*time.Second, b.Backoff(5))
		require.Equal(t, 60*time.Second, b.Backoff(6))
	})
	t.Run("Custom", func(t *testing.T) {
		b := &ArangoBackupSpecBackOff{
			MinDelay:   util.NewInt(20),
			MaxDelay:   util.NewInt(120),
			Iterations: util.NewInt(10),
		}

		require.Equal(t, 20*time.Second, b.Backoff(0))
		require.Equal(t, 30*time.Second, b.Backoff(1))
		require.Equal(t, 40*time.Second, b.Backoff(2))
		require.Equal(t, 50*time.Second, b.Backoff(3))
		require.Equal(t, 60*time.Second, b.Backoff(4))
		require.Equal(t, 70*time.Second, b.Backoff(5))
		require.Equal(t, 80*time.Second, b.Backoff(6))
		require.Equal(t, 90*time.Second, b.Backoff(7))
		require.Equal(t, 100*time.Second, b.Backoff(8))
		require.Equal(t, 110*time.Second, b.Backoff(9))
		require.Equal(t, 120*time.Second, b.Backoff(10))
		require.Equal(t, 120*time.Second, b.Backoff(11))
	})

	t.Run("Invalid", func(t *testing.T) {
		b := &ArangoBackupSpecBackOff{
			MinDelay:   util.NewInt(-1),
			MaxDelay:   util.NewInt(-1),
			Iterations: util.NewInt(0),
		}

		require.Equal(t, 0, b.GetMinDelay())
		require.Equal(t, 0, b.GetMaxDelay())
		require.Equal(t, 1, b.GetIterations())

		require.Equal(t, 0*time.Second, b.Backoff(12345))
	})

	t.Run("Max < Min", func(t *testing.T) {
		b := &ArangoBackupSpecBackOff{
			MinDelay: util.NewInt(50),
			MaxDelay: util.NewInt(20),
		}

		require.Equal(t, 20, b.GetMinDelay())
		require.Equal(t, 20, b.GetMaxDelay())
		require.Equal(t, 5, b.GetIterations())
	})
}
