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

	pbImplAuthenticationV1 "github.com/arangodb/kube-arangodb/integrations/authentication/v1"
	pbAuthorizationV1 "github.com/arangodb/kube-arangodb/integrations/authorization/v1/definition"
	pbImplAuthorizationV1Shared "github.com/arangodb/kube-arangodb/integrations/authorization/v1/shared"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
	"github.com/arangodb/kube-arangodb/pkg/util/tests"
	"github.com/arangodb/kube-arangodb/pkg/util/tests/tgrpc"
)

func NewAuthenticationHandler(t *testing.T, mods ...util.ModR[pbImplAuthenticationV1.Configuration]) (tests.TokenManager, svc.Handler) {
	directory := tests.NewTokenManager(t)

	var currentMods []util.ModR[pbImplAuthenticationV1.Configuration]

	currentMods = append(currentMods, func(c pbImplAuthenticationV1.Configuration) pbImplAuthenticationV1.Configuration {
		c.Path = directory.Path()
		return c
	})

	currentMods = append(currentMods, mods...)

	handler, err := pbImplAuthenticationV1.New(t.Context(), pbImplAuthenticationV1.NewConfiguration().With(currentMods...))
	require.NoError(t, err)

	return directory, handler
}

func Handler(plugin pbImplAuthorizationV1Shared.Plugin, mods ...util.ModR[Configuration]) svc.Handler {
	return newInternal(NewConfiguration().With(mods...), plugin)
}

func Server(t *testing.T, ctx context.Context, handlers ...svc.Handler) svc.ServiceStarter {

	local, err := svc.NewService(svc.Configuration{
		Address: "127.0.0.1:0",
		Gateway: &svc.ConfigurationGateway{
			Address: "127.0.0.1:0",
		},
	}, handlers...)
	require.NoError(t, err)

	return local.Start(ctx)
}

func Client(t *testing.T, ctx context.Context, handlers ...svc.Handler) (pbAuthorizationV1.AuthorizationV1Client, string) {
	start := Server(t, ctx, handlers...)

	client := tgrpc.NewGRPCClient(t, ctx, pbAuthorizationV1.NewAuthorizationV1Client, start.Address())

	return client, start.HTTPAddress()
}
