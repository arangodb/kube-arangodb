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

package v1

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	pbMetaV1 "github.com/arangodb/kube-arangodb/integrations/meta/v1/definition"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
	tcache "github.com/arangodb/kube-arangodb/pkg/util/tests/cache"
	"github.com/arangodb/kube-arangodb/pkg/util/tests/tgrpc"
)

func Test_Types(t *testing.T) {
	z := `{
  "meta": {
    "updatedAt": "2025-07-07T15:13:09Z"
  },
  "object": {
    "Object": {
      "type_url": "type.googleapis.com/arangodb_platform_internal.metadata_store.GenAiProjectNames",
      "value": "ChV0ZXN0X3Byb2plY3RfNmVhYWM3MjM="
    }
  }
}`

	var obj Object

	require.NoError(t, json.Unmarshal([]byte(z), &obj))

	n, err := json.Marshal(obj)
	require.NoError(t, err)

	require.JSONEq(t, z, string(n))
}

func Handler(t *testing.T, ctx context.Context, mods ...util.ModR[Configuration]) svc.Handler {
	handler, err := newInternalWithRemoteCache(ctx, NewConfiguration().With(mods...), tcache.NewRemoteCache[*Object]())
	require.NoError(t, err)

	return handler
}

func Server(t *testing.T, ctx context.Context, mods ...util.ModR[Configuration]) svc.ServiceStarter {
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
	}, Handler(t, ctx, currentMods...))
	require.NoError(t, err)

	return local.Start(ctx)
}

func Client(t *testing.T, ctx context.Context, mods ...util.ModR[Configuration]) pbMetaV1.MetaV1Client {
	start := Server(t, ctx, mods...)

	client := tgrpc.NewGRPCClient(t, ctx, pbMetaV1.NewMetaV1Client, start.Address())

	return client
}

func Test_Service(t *testing.T) {
	ctx, c := context.WithCancel(context.Background())
	defer c()

	Client(t, ctx)
}
