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

package authentication

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

// Test_applyClientAuthentication_AppendsWhenAbsent verifies the service token is stamped when the outgoing
// context carries no authorization yet (the regular integration-originated call).
func Test_applyClientAuthentication_AppendsWhenAbsent(t *testing.T) {
	ctx, err := applyClientAuthentication(context.Background(), Static("service-token"))
	require.NoError(t, err)

	md, ok := metadata.FromOutgoingContext(ctx)
	require.True(t, ok)
	require.Equal(t, []string{"bearer service-token"}, md.Get("authorization"))
}

// Test_applyClientAuthentication_PreservesForwardedCaller verifies a forwarded caller token (as set by the
// transparent proxy) is kept untouched and the service token is NOT appended on top of it - a second value
// would be rejected by the server authenticator and would discard the caller identity used for RBAC.
func Test_applyClientAuthentication_PreservesForwardedCaller(t *testing.T) {
	in := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", "bearer caller-token"))

	ctx, err := applyClientAuthentication(in, Static("service-token"))
	require.NoError(t, err)

	md, ok := metadata.FromOutgoingContext(ctx)
	require.True(t, ok)
	require.Equal(t, []string{"bearer caller-token"}, md.Get("authorization"))
}

// Test_applyClientAuthentication_NoTokenNoChange verifies that when the authenticator yields no token nothing
// is added to the outgoing context.
func Test_applyClientAuthentication_NoTokenNoChange(t *testing.T) {
	ctx, err := applyClientAuthentication(context.Background(), NewEmptyAuthentication())
	require.NoError(t, err)

	md, ok := metadata.FromOutgoingContext(ctx)
	if ok {
		require.Empty(t, md.Get("authorization"))
	}
}

func Test_hasOutgoingAuthorization(t *testing.T) {
	require.False(t, hasOutgoingAuthorization(context.Background()))
	require.False(t, hasOutgoingAuthorization(metadata.NewOutgoingContext(context.Background(), metadata.Pairs("other", "value"))))
	require.True(t, hasOutgoingAuthorization(metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", "bearer t"))))
}
