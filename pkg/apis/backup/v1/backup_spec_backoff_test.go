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

	"github.com/stretchr/testify/assert"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

func TestArangoBackupSpecBackOff_Backoff(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		var b *ArangoBackupSpecBackOff

		assert.Equal(t, 30*time.Second, b.Backoff(0))
		assert.Equal(t, 144*time.Second, b.Backoff(1))
		assert.Equal(t, 258*time.Second, b.Backoff(2))
		assert.Equal(t, 372*time.Second, b.Backoff(3))
		assert.Equal(t, 486*time.Second, b.Backoff(4))
		assert.Equal(t, 600*time.Second, b.Backoff(5))
		assert.Equal(t, 600*time.Second, b.Backoff(6))
	})
	t.Run("Custom", func(t *testing.T) {
		b := &ArangoBackupSpecBackOff{
			MinDelay:   util.NewType[int](20),
			MaxDelay:   util.NewType[int](120),
			Iterations: util.NewType[int](10),
		}

		assert.Equal(t, 20*time.Second, b.Backoff(0))
		assert.Equal(t, 30*time.Second, b.Backoff(1))
		assert.Equal(t, 40*time.Second, b.Backoff(2))
		assert.Equal(t, 50*time.Second, b.Backoff(3))
		assert.Equal(t, 60*time.Second, b.Backoff(4))
		assert.Equal(t, 70*time.Second, b.Backoff(5))
		assert.Equal(t, 80*time.Second, b.Backoff(6))
		assert.Equal(t, 90*time.Second, b.Backoff(7))
		assert.Equal(t, 100*time.Second, b.Backoff(8))
		assert.Equal(t, 110*time.Second, b.Backoff(9))
		assert.Equal(t, 120*time.Second, b.Backoff(10))
		assert.Equal(t, 120*time.Second, b.Backoff(11))
	})

	t.Run("Invalid", func(t *testing.T) {
		b := &ArangoBackupSpecBackOff{
			MinDelay:   util.NewType[int](-1),
			MaxDelay:   util.NewType[int](-1),
			Iterations: util.NewType[int](0),
		}

		assert.Equal(t, 0, b.GetMinDelay())
		assert.Equal(t, 0, b.GetMaxDelay())
		assert.Equal(t, 1, b.GetIterations())

		assert.Equal(t, 0*time.Second, b.Backoff(12345))
	})

	t.Run("Max < Min", func(t *testing.T) {
		b := &ArangoBackupSpecBackOff{
			MinDelay: util.NewType[int](50),
			MaxDelay: util.NewType[int](20),
		}

		assert.Equal(t, 20, b.GetMinDelay())
		assert.Equal(t, 20, b.GetMaxDelay())
		assert.Equal(t, 5, b.GetIterations())
	})
}
