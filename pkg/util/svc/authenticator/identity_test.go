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

package authenticator

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pbAuthorizationV1 "github.com/arangodb/kube-arangodb/integrations/authorization/v1/definition"
	pbImplAuthorizationV1Shared "github.com/arangodb/kube-arangodb/integrations/authorization/v1/shared"
	sidecarSvcAuthzTypes "github.com/arangodb/kube-arangodb/pkg/sidecar/services/authorization/types"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

// evaluator adapts a function to the authorization Evaluator interface.
func evaluator(fn func(ctx context.Context, req *pbAuthorizationV1.AuthorizationV1PermissionRequest) (*pbAuthorizationV1.AuthorizationV1PermissionResponse, error)) pbImplAuthorizationV1Shared.Evaluator {
	return pbImplAuthorizationV1Shared.PluginFunc(fn)
}

func allowEvaluator() pbImplAuthorizationV1Shared.Evaluator {
	return evaluator(func(_ context.Context, _ *pbAuthorizationV1.AuthorizationV1PermissionRequest) (*pbAuthorizationV1.AuthorizationV1PermissionResponse, error) {
		return &pbAuthorizationV1.AuthorizationV1PermissionResponse{Effect: sidecarSvcAuthzTypes.Effect_Allow}, nil
	})
}

func Test_Identity_GetUser(t *testing.T) {
	var nilID *Identity
	require.Equal(t, "", nilID.GetUser(), "nil identity has no user")
	require.Equal(t, "", (&Identity{}).GetUser(), "nil-user identity (server/superuser token) has no user")
	require.Equal(t, "alice", (&Identity{User: util.NewType("alice")}).GetUser())
}

func Test_Identity_GetGroups(t *testing.T) {
	var nilID *Identity
	require.Nil(t, nilID.GetGroups(), "nil identity has no groups")
	require.Nil(t, (&Identity{}).GetGroups())
	require.Equal(t, []string{"readers", "writers"}, (&Identity{Groups: []string{"readers", "writers"}}).GetGroups())
}

// A nil identity (no verified caller on the context) is rejected as Unauthenticated and the evaluator is
// never consulted.
func Test_Identity_EvaluatePermission_NilCallerIsUnauthenticated(t *testing.T) {
	var nilID *Identity

	consulted := false
	ev := evaluator(func(_ context.Context, _ *pbAuthorizationV1.AuthorizationV1PermissionRequest) (*pbAuthorizationV1.AuthorizationV1PermissionResponse, error) {
		consulted = true
		return &pbAuthorizationV1.AuthorizationV1PermissionResponse{Effect: sidecarSvcAuthzTypes.Effect_Allow}, nil
	})

	err := nilID.EvaluatePermission(context.Background(), ev, "authentication:CreateToken", "alice")
	require.Error(t, err)
	require.Equal(t, codes.Unauthenticated, status.Code(err))
	require.False(t, consulted, "the evaluator must not be consulted for a caller with no identity")
}

// The authorization subject is built from the verified identity (User + Groups-as-Roles), never from
// caller-controlled request input, and an Allow effect yields success. This is the security contract: the
// subject a caller is evaluated as is its authenticated identity.
func Test_Identity_EvaluatePermission_SubjectIsTheIdentity(t *testing.T) {
	id := &Identity{User: util.NewType("alice"), Groups: []string{"readers", "writers"}}

	var got *pbAuthorizationV1.AuthorizationV1PermissionRequest
	ev := evaluator(func(_ context.Context, req *pbAuthorizationV1.AuthorizationV1PermissionRequest) (*pbAuthorizationV1.AuthorizationV1PermissionResponse, error) {
		got = req
		return &pbAuthorizationV1.AuthorizationV1PermissionResponse{Effect: sidecarSvcAuthzTypes.Effect_Allow}, nil
	})

	require.NoError(t, id.EvaluatePermission(context.Background(), ev, "meta:Get", "some/resource"))
	require.NotNil(t, got, "the evaluator must be consulted")
	require.Equal(t, "alice", got.GetUser(), "subject user must be the identity user")
	require.Equal(t, []string{"readers", "writers"}, got.GetRoles(), "subject roles must be the identity groups")
	require.Equal(t, "meta:Get", got.GetAction())
	require.Equal(t, "some/resource", got.GetResource())
}

// A superuser (nil-user) identity is still a verified caller: it is evaluated (with no user/roles), not
// short-circuited as Unauthenticated.
func Test_Identity_EvaluatePermission_SuperuserIsEvaluated(t *testing.T) {
	id := &Identity{}

	var got *pbAuthorizationV1.AuthorizationV1PermissionRequest
	ev := evaluator(func(_ context.Context, req *pbAuthorizationV1.AuthorizationV1PermissionRequest) (*pbAuthorizationV1.AuthorizationV1PermissionResponse, error) {
		got = req
		return &pbAuthorizationV1.AuthorizationV1PermissionResponse{Effect: sidecarSvcAuthzTypes.Effect_Allow}, nil
	})

	require.NoError(t, id.EvaluatePermission(context.Background(), ev, "a", "r"))
	require.NotNil(t, got)
	require.Empty(t, got.GetUser())
	require.Empty(t, got.GetRoles())
}

// A Deny effect maps to PermissionDenied and propagates the evaluator message.
func Test_Identity_EvaluatePermission_DenyIsPermissionDenied(t *testing.T) {
	id := &Identity{User: util.NewType("alice")}

	ev := evaluator(func(_ context.Context, _ *pbAuthorizationV1.AuthorizationV1PermissionRequest) (*pbAuthorizationV1.AuthorizationV1PermissionResponse, error) {
		return &pbAuthorizationV1.AuthorizationV1PermissionResponse{Effect: sidecarSvcAuthzTypes.Effect_Deny, Message: "not allowed"}, nil
	})

	err := id.EvaluatePermission(context.Background(), ev, "a", "r")
	require.Error(t, err)
	require.Equal(t, codes.PermissionDenied, status.Code(err))
	require.Contains(t, status.Convert(err).Message(), "not allowed")
}

// A transport/plugin error from the evaluator surfaces as Internal (not as an allow).
func Test_Identity_EvaluatePermission_EvaluatorErrorIsInternal(t *testing.T) {
	id := &Identity{User: util.NewType("alice")}

	ev := evaluator(func(_ context.Context, _ *pbAuthorizationV1.AuthorizationV1PermissionRequest) (*pbAuthorizationV1.AuthorizationV1PermissionResponse, error) {
		return nil, errors.New("evaluator unavailable")
	})

	err := id.EvaluatePermission(context.Background(), ev, "a", "r")
	require.Error(t, err)
	require.Equal(t, codes.Internal, status.Code(err))
}

// WithIdentity / GetIdentity round-trip; a context carrying no identity yields nil.
func Test_Identity_Context_RoundTrip(t *testing.T) {
	require.Nil(t, GetIdentity(context.Background()), "a context with no identity resolves to nil")

	id := &Identity{User: util.NewType("alice"), Groups: []string{"g"}}
	ctx := WithIdentity(context.Background(), id)
	require.Same(t, id, GetIdentity(ctx))

	// A nil identity stored on the context round-trips as a nil *Identity, which EvaluatePermission rejects.
	ctxNil := WithIdentity(context.Background(), nil)
	require.Nil(t, GetIdentity(ctxNil))
	require.Equal(t, codes.Unauthenticated, status.Code(GetIdentity(ctxNil).EvaluatePermission(ctxNil, allowEvaluator(), "a", "r")))
}
