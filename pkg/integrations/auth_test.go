//
// DISCLAIMER
//
// Copyright 2024-2026 ArangoDB GmbH, Cologne, Germany
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

package integrations

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"

	"github.com/arangodb/kube-arangodb/pkg/util/shutdown"
	"github.com/arangodb/kube-arangodb/pkg/util/tests/tgrpc"
)

func Test_AuthCases(t *testing.T) {
	c, health, internal, external := startService(t,
		"--health.auth.type=None",
		"--services.external.auth.token=test1",
		"--services.external.auth.type=Token",
		"--services.auth.token=test2",
		"--services.auth.type=Token",
	)
	defer c.Require(t)

	t.Run("Without auth", func(t *testing.T) {
		t.Run("health", func(t *testing.T) {
			require.NoError(t, executeSync(t, shutdown.Context(),
				fmt.Sprintf("--address=127.0.0.1:%d", health),
				"--token=",
				"client",
				"health",
				"v1"))
		})
		t.Run("internal", func(t *testing.T) {
			tgrpc.AsGRPCError(t, executeSync(t, shutdown.Context(),
				fmt.Sprintf("--address=127.0.0.1:%d", internal),
				"--token=",
				"client",
				"health",
				"v1")).
				Code(t, codes.Unauthenticated).
				Errorf(t, "Unauthorized")
		})
		t.Run("external", func(t *testing.T) {
			tgrpc.AsGRPCError(t, executeSync(t, shutdown.Context(),
				fmt.Sprintf("--address=127.0.0.1:%d", external),
				"--token=",
				"client",
				"health",
				"v1")).
				Code(t, codes.Unauthenticated).
				Errorf(t, "Unauthorized")
		})
	})

	t.Run("With auth 1", func(t *testing.T) {
		t.Run("health", func(t *testing.T) {
			require.NoError(t, executeSync(t, shutdown.Context(),
				fmt.Sprintf("--address=127.0.0.1:%d", health),
				"--token=test1",
				"client",
				"health",
				"v1"))
		})
		t.Run("internal", func(t *testing.T) {
			tgrpc.AsGRPCError(t, executeSync(t, shutdown.Context(),
				fmt.Sprintf("--address=127.0.0.1:%d", internal),
				"--token=test1",
				"client",
				"health",
				"v1")).
				Code(t, codes.Unauthenticated).
				Errorf(t, "Unauthorized")
		})
		t.Run("external", func(t *testing.T) {
			require.NoError(t, executeSync(t, shutdown.Context(),
				fmt.Sprintf("--address=127.0.0.1:%d", external),
				"--token=test1",
				"client",
				"health",
				"v1"))
		})
	})

	t.Run("With auth 2", func(t *testing.T) {
		t.Run("health", func(t *testing.T) {
			require.NoError(t, executeSync(t, shutdown.Context(),
				fmt.Sprintf("--address=127.0.0.1:%d", health),
				"--token=test2",
				"client",
				"health",
				"v1"))
		})
		t.Run("internal", func(t *testing.T) {
			require.NoError(t, executeSync(t, shutdown.Context(),
				fmt.Sprintf("--address=127.0.0.1:%d", internal),
				"--token=test2",
				"client",
				"health",
				"v1"))
		})
		t.Run("external", func(t *testing.T) {
			tgrpc.AsGRPCError(t, executeSync(t, shutdown.Context(),
				fmt.Sprintf("--address=127.0.0.1:%d", external),
				"--token=test2",
				"client",
				"health",
				"v1")).
				Code(t, codes.Unauthenticated).
				Errorf(t, "Unauthorized")
		})
	})
}
