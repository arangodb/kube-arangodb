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

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"

	pbAuthorizationV1 "github.com/arangodb/kube-arangodb/integrations/authorization/v1/definition"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

// Test_Evaluate_RefreshHeader verifies that the RefreshHeader request metadata forces a pool refresh
// before evaluation (read-your-writes consistency), and that evaluation without it never refreshes.
func Test_Evaluate_RefreshHeader(t *testing.T) {
	ctx, c := context.WithCancel(context.Background())
	defer c()

	p := newPluginTest()
	client, _ := Client(t, ctx, Handler(p))

	req := &pbAuthorizationV1.AuthorizationV1PermissionRequest{
		User:     util.NewType("admin"),
		Action:   "test:Get",
		Resource: "test",
	}

	t.Run("without header does not refresh", func(t *testing.T) {
		before := p.RefreshCount()
		_, err := client.Evaluate(ctx, req)
		require.NoError(t, err)
		require.Equal(t, before, p.RefreshCount(), "Evaluate must not refresh when the header is absent")
	})

	t.Run("with header refreshes before evaluating", func(t *testing.T) {
		before := p.RefreshCount()
		ctxR := metadata.AppendToOutgoingContext(ctx, pbAuthorizationV1.RefreshHeader, "true")
		_, err := client.Evaluate(ctxR, req)
		require.NoError(t, err)
		require.Equal(t, before+1, p.RefreshCount(), "Evaluate must refresh once when the header is set")
	})

	t.Run("header value other than true does not refresh", func(t *testing.T) {
		before := p.RefreshCount()
		ctxR := metadata.AppendToOutgoingContext(ctx, pbAuthorizationV1.RefreshHeader, "false")
		_, err := client.Evaluate(ctxR, req)
		require.NoError(t, err)
		require.Equal(t, before, p.RefreshCount())
	})
}
