//
// DISCLAIMER
//
// Copyright 2025 ArangoDB GmbH, Cologne, Germany
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

package kclient

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/shutdown"
)

func waitForExecution(ctx context.Context, count int, rl *rateLimiter) error {
	return util.ParallelProcessErr(func(in int) error {
		return rl.Wait(ctx)
	}, 16, util.IntInput(count))
}

func Test_RateLimiter(t *testing.T) {
	rl := getRateLimiter("TEST1")

	rl.setQPS(1)
	rl.setBurst(1)

	s := time.Now()

	require.NoError(t, waitForExecution(shutdown.Context(), 3, rl))

	require.True(t, time.Since(s) > 2*time.Second)
	require.True(t, time.Since(s) < 3*time.Second)
}

func Test_RateLimiter_Multi(t *testing.T) {
	rl := getRateLimiter("TEST2")

	rl.setQPS(128)
	rl.setBurst(1)

	s := time.Now()

	require.NoError(t, waitForExecution(shutdown.Context(), 200, rl))

	require.True(t, time.Since(s) > 1*time.Second)
	require.True(t, time.Since(s) < 2*time.Second)
}

func Test_RateLimiter_Multi_Large(t *testing.T) {
	rl := getRateLimiter("TEST2L")

	rl.setQPS(128)
	rl.setBurst(1)

	s := time.Now()

	require.NoError(t, waitForExecution(shutdown.Context(), 257, rl))

	require.True(t, time.Since(s) > 2*time.Second)
	require.True(t, time.Since(s) < 3*time.Second)
}

func Test_RateLimiter_EnsureErrorTrack(t *testing.T) {
	rl := getRateLimiter("TEST3")

	rl.setQPS(1)
	rl.setBurst(1)

	ctx, cancel := context.WithTimeout(shutdown.Context(), time.Second)
	defer cancel()

	require.NoError(t, rl.Wait(ctx))
	require.Error(t, rl.Wait(ctx))
}
