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

package v1

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	pbLinkV1 "github.com/arangodb/kube-arangodb/integrations/link/v1/definition"
)

func createJobWithPriority(t *testing.T, impl *implementation, input string, priority int32) string {
	t.Helper()

	resp, err := impl.CreateJob(context.Background(), &pbLinkV1.CreateJobRequest{
		Input:    []byte(input),
		Priority: priority,
	})
	require.NoError(t, err)
	require.NotEmpty(t, resp.Id)
	return resp.Id
}

// Test_PickUp_PriorityOrder verifies that higher-priority jobs are picked up first.
func Test_PickUp_PriorityOrder(t *testing.T) {
	impl := newTestImpl(t)

	low := createJobWithPriority(t, impl, "low", 1)
	high := createJobWithPriority(t, impl, "high", 100)
	medium := createJobWithPriority(t, impl, "medium", 10)

	require.Equal(t, high, pickUp(t, impl), "highest priority first")
	require.Equal(t, medium, pickUp(t, impl), "medium priority second")
	require.Equal(t, low, pickUp(t, impl), "lowest priority last")
}

// Test_PickUp_FIFOWithinPriority verifies that jobs of equal priority are picked up
// oldest-first (FIFO), driven by the creation epoch embedded in the job id.
func Test_PickUp_FIFOWithinPriority(t *testing.T) {
	impl := newTestImpl(t)

	// Sleep between creations so the epoch embedded in each id is strictly increasing,
	// making the FIFO order deterministic (ties within a nanosecond fall back to the uuid).
	first := createJobWithPriority(t, impl, "first", 5)
	time.Sleep(2 * time.Millisecond)
	second := createJobWithPriority(t, impl, "second", 5)
	time.Sleep(2 * time.Millisecond)
	third := createJobWithPriority(t, impl, "third", 5)

	require.Equal(t, first, pickUp(t, impl))
	require.Equal(t, second, pickUp(t, impl))
	require.Equal(t, third, pickUp(t, impl))
}

// Test_PickUp_PriorityBeatsAge verifies priority wins over creation order.
func Test_PickUp_PriorityBeatsAge(t *testing.T) {
	impl := newTestImpl(t)

	old := createJobWithPriority(t, impl, "old-low", 0)
	newer := createJobWithPriority(t, impl, "new-high", 50)

	require.Equal(t, newer, pickUp(t, impl), "newer high-priority job jumps the older low-priority one")
	require.Equal(t, old, pickUp(t, impl))
}
