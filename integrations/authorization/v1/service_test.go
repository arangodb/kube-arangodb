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
	"fmt"
	goHttp "net/http"
	"testing"

	"github.com/stretchr/testify/require"

	pbAuthorizationV1 "github.com/arangodb/kube-arangodb/integrations/authorization/v1/definition"
	pbImplAuthorizationV1Shared "github.com/arangodb/kube-arangodb/integrations/authorization/v1/shared"
	"github.com/arangodb/kube-arangodb/pkg/util"
	ugrpc "github.com/arangodb/kube-arangodb/pkg/util/grpc"
	"github.com/arangodb/kube-arangodb/pkg/util/http"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
	"github.com/arangodb/kube-arangodb/pkg/util/tests/tgrpc"
)

func Handler(plugin pbImplAuthorizationV1Shared.Plugin, mods ...util.ModR[Configuration]) svc.Handler {
	return newInternal(NewConfiguration().With(mods...), plugin)
}

func Server(t *testing.T, ctx context.Context, plugin pbImplAuthorizationV1Shared.Plugin, mods ...util.ModR[Configuration]) svc.ServiceStarter {
	var currentMods []util.ModR[Configuration]

	currentMods = append(currentMods, func(c Configuration) Configuration {
		return c
	})

	currentMods = append(currentMods, mods...)

	local, err := svc.NewService(svc.Configuration{
		Address: "127.0.0.1:0",
		Gateway: &svc.ConfigurationGateway{
			Address: "127.0.0.1:0",
		},
	}, Handler(plugin, currentMods...))
	require.NoError(t, err)

	return local.Start(ctx)
}

func Client(t *testing.T, ctx context.Context, plugin pbImplAuthorizationV1Shared.Plugin, mods ...util.ModR[Configuration]) (pbAuthorizationV1.AuthorizationV1Client, string) {
	start := Server(t, ctx, plugin, mods...)

	client := tgrpc.NewGRPCClient(t, ctx, pbAuthorizationV1.NewAuthorizationV1Client, start.Address())

	return client, start.HTTPAddress()
}

func Test_Service(t *testing.T) {
	ctx, c := context.WithCancel(context.Background())
	defer c()

	p := newPluginTest()

	client, _ := Client(t, ctx, p)

	resp, err := client.Evaluate(ctx, &pbAuthorizationV1.AuthorizationV1PermissionRequest{
		User:     "admin",
		Action:   "test:Get",
		Resource: "test",
	})
	require.NoError(t, err)
	require.EqualValues(t, pbAuthorizationV1.AuthorizationV1Effect_Allow, resp.GetEffect())
}

func Test_ServiceHTTP(t *testing.T) {
	ctx, c := context.WithCancel(context.Background())
	defer c()

	p := newPluginTest()

	_, endpoint := Client(t, ctx, p)

	data, err := http.Post[ugrpc.Object[*pbAuthorizationV1.AuthorizationV1PermissionRequest], any, error](ctx, goHttp.DefaultClient, ugrpc.NewObject(&pbAuthorizationV1.AuthorizationV1PermissionRequest{
		User:     "admin",
		Action:   "test:Get",
		Resource: "test",
	}), fmt.Sprintf("http://%s/_integration/authorization/v1/evaluate", endpoint)).WithCode(200).Data()
	require.NoError(t, err)

	t.Logf("Response: %s", string(data))
}
