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

package authorization

import (
	"context"
	"fmt"
	goStrings "strings"
	"testing"
	"time"

	"github.com/dchest/uniuri"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/arangodb/go-driver/v2/arangodb"

	sidecarSvcAuthzClient "github.com/arangodb/kube-arangodb/pkg/sidecar/services/authorization/client"
	sidecarSvcAuthz "github.com/arangodb/kube-arangodb/pkg/sidecar/services/authorization/definition"
	sidecarSvcAuthzTypes "github.com/arangodb/kube-arangodb/pkg/sidecar/services/authorization/types"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod/db"
	"github.com/arangodb/kube-arangodb/pkg/util/cache"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
	"github.com/arangodb/kube-arangodb/pkg/util/svc/authentication"
	"github.com/arangodb/kube-arangodb/pkg/util/svc/authenticator"
	"github.com/arangodb/kube-arangodb/pkg/util/tests"
	"github.com/arangodb/kube-arangodb/pkg/util/tests/tgrpc"
	utilToken "github.com/arangodb/kube-arangodb/pkg/util/token"
)

func Handler(t *testing.T) svc.Handler {
	return NewAuthorizer(db.NewClient(cache.NewObject(tests.TestArangoDBConfig(t).ClientCache())).
		CreateDatabase(fmt.Sprintf("db-%s", goStrings.ToLower(uniuri.NewLen(8))), &arangodb.CreateDatabaseOptions{}).
		CreateCollection("_users", db.StaticProps(arangodb.CreateCollectionPropertiesV2{
			IsSystem: util.NewType(true),
		})).Database())
}

func Server(t *testing.T, ctx context.Context, tm tests.TokenManager) svc.ServiceStarter {
	local, err := svc.NewService(svc.Configuration{
		Authenticator: authenticator.Required(authenticator.NewJWTAuthentication(tm.Path())),
		Address:       "127.0.0.1:0",
		Gateway: &svc.ConfigurationGateway{
			Address: "127.0.0.1:0",
		},
	}, Handler(t))
	require.NoError(t, err)

	return local.Start(ctx)
}

func Client(t *testing.T, ctx context.Context, tm tests.TokenManager, auth authentication.Authentication) (sidecarSvcAuthz.AuthorizationPoolServiceClient, sidecarSvcAuthz.AuthorizationAPIClient) {
	start := Server(t, ctx, tm)

	var opts []grpc.DialOption

	opts = append(opts, authentication.NewInterceptorClientOptions(auth)...)

	return tgrpc.NewGRPCClient(t, ctx, sidecarSvcAuthz.NewAuthorizationPoolServiceClient, start.Address(), opts...), tgrpc.NewGRPCClient(t, ctx, sidecarSvcAuthz.NewAuthorizationAPIClient, start.Address(), opts...)
}

func Test_Service(t *testing.T) {
	ctx, c := context.WithCancel(t.Context())
	defer c()

	tm := tests.NewTokenManager(t)

	q, api := Client(t, ctx, tm, authentication.NewCachedAuthentication(cache.NewObject(tm.TokenSignature(t, utilToken.WithRelativeDuration(time.Second)))))

	token := tests.GenerateJWTToken()

	tm.Set(t, token)

	z := sidecarSvcAuthzClient.NewClient(t.Context(), q)

	zctx, c := context.WithTimeout(t.Context(), time.Second)
	defer c()

	require.True(t, z.Wait(zctx))

	tgrpc.NewExecutor(t, api.APICreatePolicy, &sidecarSvcAuthz.AuthorizationAPIPolicyRequest{
		Name: "example",
		Item: nil,
	}).Code(t, codes.InvalidArgument)

	time.Sleep(5 * time.Second)

	tgrpc.NewExecutor(t, api.APICreatePolicy, &sidecarSvcAuthz.AuthorizationAPIPolicyRequest{
		Name: "example",
		Item: &sidecarSvcAuthzTypes.Policy{},
	}).Code(t, codes.OK)

	time.Sleep(5 * time.Second)

	tgrpc.NewExecutor(t, api.APICreatePolicy, &sidecarSvcAuthz.AuthorizationAPIPolicyRequest{
		Name: "example2",
		Item: &sidecarSvcAuthzTypes.Policy{},
	}).Code(t, codes.OK)

	tgrpc.NewExecutor(t, api.APICreatePolicy, &sidecarSvcAuthz.AuthorizationAPIPolicyRequest{
		Name: "example3",
		Item: &sidecarSvcAuthzTypes.Policy{},
	}).Code(t, codes.OK)

	tgrpc.NewExecutor(t, api.APICreatePolicy, &sidecarSvcAuthz.AuthorizationAPIPolicyRequest{
		Name: "example4",
		Item: &sidecarSvcAuthzTypes.Policy{},
	}).Code(t, codes.OK)

	tgrpc.NewExecutor(t, api.APICreatePolicy, &sidecarSvcAuthz.AuthorizationAPIPolicyRequest{
		Name: "example5",
		Item: &sidecarSvcAuthzTypes.Policy{},
	}).Code(t, codes.OK)

	time.Sleep(5 * time.Minute)
}
