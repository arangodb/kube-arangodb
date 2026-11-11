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

package v1

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"

	pbAuthenticationV1 "github.com/arangodb/kube-arangodb/integrations/authentication/v1/definition"
	pbAuthorizationV1 "github.com/arangodb/kube-arangodb/integrations/authorization/v1/definition"
	pbImplAuthorizationV1Shared "github.com/arangodb/kube-arangodb/integrations/authorization/v1/shared"
	sidecarSvcAuthzTypes "github.com/arangodb/kube-arangodb/pkg/sidecar/services/authorization/types"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/cache"
	utilConstants "github.com/arangodb/kube-arangodb/pkg/util/constants"
	utilConstantsContext "github.com/arangodb/kube-arangodb/pkg/util/constants/context"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
	"github.com/arangodb/kube-arangodb/pkg/util/svc/authenticator"
	"github.com/arangodb/kube-arangodb/pkg/util/tests"
	"github.com/arangodb/kube-arangodb/pkg/util/tests/tgrpc"
	utilToken "github.com/arangodb/kube-arangodb/pkg/util/token"
)

func Handler(t *testing.T, ctx context.Context, mods ...util.ModR[Configuration]) svc.Handler {
	handler, err := New(ctx, NewConfiguration().With(mods...))
	require.NoError(t, err)

	return handler
}

func Server(t *testing.T, ctx context.Context, mods ...util.ModR[Configuration]) (tests.TokenManager, svc.ServiceStarter) {
	tm := tests.NewTokenManager(t)

	ctx = utilConstantsContext.ArangoDBClientCache.Set(ctx, cache.NewObject(tests.TestArangoClientCacheFunc(t)))

	var currentMods []util.ModR[Configuration]

	currentMods = append(currentMods, func(c Configuration) Configuration {
		c.Path = tm.Path()
		return c
	})

	currentMods = append(currentMods, mods...)

	local, err := svc.NewService(svc.Configuration{
		Address: "127.0.0.1:0",
		Gateway: &svc.ConfigurationGateway{
			Address: "127.0.0.1:0",
		},
	}, Handler(t, ctx, currentMods...))
	require.NoError(t, err)

	return tm, local.Start(ctx)
}

func Client(t *testing.T, ctx context.Context, mods ...util.ModR[Configuration]) (pbAuthenticationV1.AuthenticationV1Client, tests.TokenManager) {
	directory, start := Server(t, ctx, mods...)

	client := tgrpc.NewGRPCClient(t, ctx, pbAuthenticationV1.NewAuthenticationV1Client, start.Address())

	return client, directory
}

func Test_Service(t *testing.T) {
	ctx, c := context.WithCancel(context.Background())
	defer c()

	client, directory := Client(t, ctx)

	directory.Activate(t, tests.GenerateJWTToken())

	token, err := client.CreateToken(ctx, &pbAuthenticationV1.CreateTokenRequest{})
	require.NoError(t, err)

	t.Logf("Token generated for user %s: %s", token.GetUser(), token.Token)

	require.EqualValues(t, "root", token.GetUser())

	valid, err := client.Validate(ctx, &pbAuthenticationV1.ValidateRequest{
		Token: token.Token,
	})
	require.NoError(t, err)

	require.True(t, valid.IsValid)
	require.NotNil(t, valid.Details)
	require.EqualValues(t, token.User, valid.Details.User)
}

func Test_Service_DifferentDefaultUser(t *testing.T) {
	ctx, c := context.WithCancel(context.Background())
	defer c()

	client, directory := Client(t, ctx, func(c Configuration) Configuration {
		c.Create.DefaultUser = "different"
		return c
	})

	directory.Set(t, tests.GenerateJWTToken())

	token, err := client.CreateToken(ctx, &pbAuthenticationV1.CreateTokenRequest{})
	require.NoError(t, err)

	t.Logf("Token generated for user %s: %s", token.GetUser(), token.Token)

	require.EqualValues(t, "different", token.GetUser())

	valid, err := client.Validate(ctx, &pbAuthenticationV1.ValidateRequest{
		Token: token.Token,
	})
	require.NoError(t, err)

	require.True(t, valid.IsValid)
	require.NotNil(t, valid.Details)
	require.EqualValues(t, token.User, valid.Details.User)
}

func Test_Service_TokenForDefaultUser(t *testing.T) {
	ctx, c := context.WithCancel(context.Background())
	defer c()

	client, directory := Client(t, ctx)

	directory.Set(t, tests.GenerateJWTToken())

	token, err := client.CreateToken(ctx, &pbAuthenticationV1.CreateTokenRequest{})
	require.NoError(t, err)

	t.Logf("Token generated for user %s: %s", token.GetUser(), token.Token)

	require.EqualValues(t, "root", token.GetUser())

	valid, err := client.Validate(ctx, &pbAuthenticationV1.ValidateRequest{
		Token: token.Token,
	})
	require.NoError(t, err)

	require.True(t, valid.IsValid)
	require.NotNil(t, valid.Details)
	require.EqualValues(t, token.User, valid.Details.User)
}

func Test_Service_TokenForNamedUser(t *testing.T) {
	ctx, c := context.WithCancel(context.Background())
	defer c()

	client, directory := Client(t, ctx)

	directory.Set(t, tests.GenerateJWTToken())

	token, err := client.CreateToken(ctx, &pbAuthenticationV1.CreateTokenRequest{
		User: util.NewType("other"),
	})
	require.NoError(t, err)

	t.Logf("Token generated for user %s: %s", token.GetUser(), token.Token)

	require.EqualValues(t, "other", token.GetUser())

	valid, err := client.Validate(ctx, &pbAuthenticationV1.ValidateRequest{
		Token: token.Token,
	})
	require.NoError(t, err)

	require.True(t, valid.IsValid)
	require.NotNil(t, valid.Details)
	require.EqualValues(t, token.User, valid.Details.User)
}

func symmetricSecret(t *testing.T) utilToken.Secret {
	s := utilToken.NewJWTSecret([]byte("0123456789abcdef0123456789abcdef"))
	require.Empty(t, s.PublicKey())
	return s
}

func asymmetricSecret(t *testing.T) utilToken.Secret {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	s, err := utilToken.NewECDSASignSecret(key)
	require.NoError(t, err)
	require.NotEmpty(t, s.PublicKey())
	return s
}

const authzTestUser = "user"

func tokenRequestFor() *pbAuthenticationV1.CreateTokenRequest {
	return &pbAuthenticationV1.CreateTokenRequest{User: util.NewType(authzTestUser), Groups: []string{"g"}}
}

// callerCtx returns a context carrying a verified caller identity, as the server auth interceptor would
// set it. The authorization subject is this identity, never the request.
func callerCtx(user string, groups ...string) context.Context {
	return authenticator.WithIdentity(context.Background(), &authenticator.Identity{User: util.NewType(user), Groups: groups})
}

// superuserCtx returns a context carrying a superuser identity (nil user), as a server/superuser token
// without a preferred_username claim produces; the SuperUser wrapper allows it.
func superuserCtx() context.Context {
	return authenticator.WithIdentity(context.Background(), &authenticator.Identity{})
}

// Without central services the IAM check is skipped entirely, even with a denying evaluator.
func Test_authorizeCreateToken_SkippedWithoutCentral(t *testing.T) {
	i := &implementation{authz: pbImplAuthorizationV1Shared.NewNeverPlugin()}
	require.NoError(t, i.authorizeCreateToken(context.Background(), asymmetricSecret(t), tokenRequestFor(), authzTestUser))
}

// With central services but a symmetric signing key there is no remote-validation path, so the IAM
// check is skipped even with a denying evaluator.
func Test_authorizeCreateToken_SkippedWithSymmetricKey(t *testing.T) {
	t.Setenv(string(utilConstants.CENTRAL_INTEGRATION_SERVICE_ADDRESS), "127.0.0.1:0")

	i := &implementation{authz: pbImplAuthorizationV1Shared.NewNeverPlugin()}
	require.NoError(t, i.authorizeCreateToken(context.Background(), symmetricSecret(t), tokenRequestFor(), authzTestUser))
}

// With central services + asymmetric key, a caller the authorization service denies is rejected with
// PermissionDenied.
func Test_authorizeCreateToken_DeniesUnauthorizedUser(t *testing.T) {
	t.Setenv(string(utilConstants.CENTRAL_INTEGRATION_SERVICE_ADDRESS), "127.0.0.1:0")

	i := &implementation{authz: pbImplAuthorizationV1Shared.NewNeverPlugin()}

	err := i.authorizeCreateToken(callerCtx(authzTestUser), asymmetricSecret(t), tokenRequestFor(), authzTestUser)
	require.Error(t, err)

	s, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.PermissionDenied, s.Code())
}

// With central services + asymmetric key, a caller the authorization service allows passes.
func Test_authorizeCreateToken_AllowsAuthorizedUser(t *testing.T) {
	t.Setenv(string(utilConstants.CENTRAL_INTEGRATION_SERVICE_ADDRESS), "127.0.0.1:0")

	i := &implementation{authz: pbImplAuthorizationV1Shared.NewAlwaysPlugin()}
	require.NoError(t, i.authorizeCreateToken(callerCtx(authzTestUser), asymmetricSecret(t), tokenRequestFor(), authzTestUser))
}

// With central services + asymmetric key, a request carrying no verified caller identity is rejected as
// Unauthenticated even when the evaluator would allow everything: the subject must come from the verified
// identity, not from the (caller-controlled) request.
func Test_authorizeCreateToken_RequiresAuthenticatedCaller(t *testing.T) {
	t.Setenv(string(utilConstants.CENTRAL_INTEGRATION_SERVICE_ADDRESS), "127.0.0.1:0")

	i := &implementation{authz: pbImplAuthorizationV1Shared.NewAlwaysPlugin()}

	err := i.authorizeCreateToken(context.Background(), asymmetricSecret(t), tokenRequestFor(), authzTestUser)
	require.Error(t, err)

	s, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.Unauthenticated, s.Code())
}

// A superuser caller (nil-user identity) is a privileged mint: the SuperUser wrapper allows it even when
// the underlying evaluator denies everything.
func Test_authorizeCreateToken_SuperuserForDefaultUser(t *testing.T) {
	t.Setenv(string(utilConstants.CENTRAL_INTEGRATION_SERVICE_ADDRESS), "127.0.0.1:0")

	i := &implementation{authz: pbImplAuthorizationV1Shared.SuperUser(pbImplAuthorizationV1Shared.NewNeverPlugin())}
	require.NoError(t, i.authorizeCreateToken(superuserCtx(), asymmetricSecret(t), &pbAuthenticationV1.CreateTokenRequest{}, "root"))
}

func tokenRequestForGroups(groups ...string) *pbAuthenticationV1.CreateTokenRequest {
	return &pbAuthenticationV1.CreateTokenRequest{User: util.NewType(authzTestUser), Groups: groups}
}

// groupEvaluator allows every evaluation except an authentication:CreateTokenGroup check whose resource
// is one of the denied groups. It records the group resources it evaluated.
func groupEvaluator(evaluated *[]string, deniedGroups ...string) pbImplAuthorizationV1Shared.Evaluator {
	denied := map[string]bool{}
	for _, g := range deniedGroups {
		denied[g] = true
	}

	return pbImplAuthorizationV1Shared.PluginFunc(func(ctx context.Context, req *pbAuthorizationV1.AuthorizationV1PermissionRequest) (*pbAuthorizationV1.AuthorizationV1PermissionResponse, error) {
		if req.GetAction() == ActionCreateTokenGroup {
			if evaluated != nil {
				*evaluated = append(*evaluated, req.GetResource())
			}
			if denied[req.GetResource()] {
				return &pbAuthorizationV1.AuthorizationV1PermissionResponse{Effect: sidecarSvcAuthzTypes.Effect_Deny, Message: "group denied"}, nil
			}
		}
		return &pbAuthorizationV1.AuthorizationV1PermissionResponse{Effect: sidecarSvcAuthzTypes.Effect_Allow}, nil
	})
}

// With central services + asymmetric key, a token mint the service allows is still rejected when the
// caller is not authorized to assign one of the requested groups.
func Test_authorizeCreateToken_DeniesUnauthorizedGroup(t *testing.T) {
	t.Setenv(string(utilConstants.CENTRAL_INTEGRATION_SERVICE_ADDRESS), "127.0.0.1:0")

	var evaluated []string
	i := &implementation{authz: groupEvaluator(&evaluated, "admin")}

	err := i.authorizeCreateToken(callerCtx(authzTestUser), asymmetricSecret(t), tokenRequestForGroups("reader", "admin"), authzTestUser)
	require.Error(t, err)

	s, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.PermissionDenied, s.Code())
	require.Contains(t, evaluated, "admin")
}

// With central services + asymmetric key, a mint whose groups are all permitted passes, and every group
// is evaluated with the dedicated group action.
func Test_authorizeCreateToken_AllowsAuthorizedGroups(t *testing.T) {
	t.Setenv(string(utilConstants.CENTRAL_INTEGRATION_SERVICE_ADDRESS), "127.0.0.1:0")

	var evaluated []string
	i := &implementation{authz: groupEvaluator(&evaluated)}

	require.NoError(t, i.authorizeCreateToken(callerCtx(authzTestUser), asymmetricSecret(t), tokenRequestForGroups("reader", "writer"), authzTestUser))
	require.ElementsMatch(t, []string{"reader", "writer"}, evaluated)
}

// A request with no groups performs no group checks.
func Test_authorizeCreateToken_NoGroupsNoGroupChecks(t *testing.T) {
	t.Setenv(string(utilConstants.CENTRAL_INTEGRATION_SERVICE_ADDRESS), "127.0.0.1:0")

	var evaluated []string
	i := &implementation{authz: groupEvaluator(&evaluated)}

	require.NoError(t, i.authorizeCreateToken(callerCtx(authzTestUser), asymmetricSecret(t), tokenRequestForGroups(), authzTestUser))
	require.Empty(t, evaluated)
}

// New builds a central-delegating evaluator when central services are enabled, without dialing eagerly.
func Test_New_BuildsCentralEvaluator(t *testing.T) {
	t.Setenv(string(utilConstants.CENTRAL_INTEGRATION_SERVICE_ADDRESS), "127.0.0.1:0")

	ctx, c := context.WithCancel(context.Background())
	defer c()

	ctx = utilConstantsContext.ArangoDBClientCache.Set(ctx, cache.NewObject(tests.TestArangoClientCacheFunc(t)))

	i, err := newInternal(ctx, NewConfiguration().With(func(c Configuration) Configuration {
		c.Path = tests.NewTokenManager(t).Path()
		return c
	}))
	require.NoError(t, err)
	require.NotNil(t, i.authz)
}

func Test_Service_WithTTL(t *testing.T) {
	ctx, c := context.WithCancel(context.Background())
	defer c()

	client, directory := Client(t, ctx)

	directory.Set(t, tests.GenerateJWTToken())

	extract := func(t *testing.T, duration time.Duration) (time.Duration, time.Duration) {
		token, err := client.CreateToken(ctx, &pbAuthenticationV1.CreateTokenRequest{
			Lifetime: durationpb.New(duration),
		})
		require.NoError(t, err)

		valid, err := client.Validate(ctx, &pbAuthenticationV1.ValidateRequest{
			Token: token.Token,
		})
		require.NoError(t, err)

		require.NotNil(t, token.Lifetime)
		require.True(t, valid.IsValid)
		require.NotNil(t, valid.Details)

		return token.Lifetime.AsDuration(), valid.Details.Lifetime.AsDuration()
	}

	t.Run("10h", func(t *testing.T) {
		base, current := extract(t, 10*time.Hour)
		require.EqualValues(t, time.Hour, base)
		require.True(t, base-time.Second < current)
	})

	t.Run("1h", func(t *testing.T) {
		base, current := extract(t, time.Hour)
		require.EqualValues(t, time.Hour, base)
		require.True(t, base-time.Second < current)
	})

	t.Run("1min", func(t *testing.T) {
		base, current := extract(t, time.Minute)
		require.EqualValues(t, time.Minute, base)
		require.True(t, base-time.Second < current)
	})

	t.Run("1sec", func(t *testing.T) {
		base, current := extract(t, time.Second)
		require.EqualValues(t, time.Minute, base)
		require.True(t, base-time.Second < current)
	})
}
