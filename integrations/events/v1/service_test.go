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
	"testing"

	"github.com/stretchr/testify/require"

	pbEventsV1 "github.com/arangodb/kube-arangodb/integrations/events/v1/definition"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
	"github.com/arangodb/kube-arangodb/pkg/util/tests/tgrpc"
)

func Handler(remote RemoteStore[*pbEventsV1.Event], mods ...util.ModR[Configuration]) svc.Handler {
	return newInternal(NewConfiguration().With(mods...), remote)
}

func Server(t *testing.T, ctx context.Context, mods ...util.ModR[Configuration]) (svc.ServiceStarter, TestRemoteStore[*pbEventsV1.Event]) {
	store := NewArangoTestStore[*pbEventsV1.Event]()

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
	}, Handler(store, currentMods...))
	require.NoError(t, err)

	return local.Start(ctx), store
}

func Client(t *testing.T, ctx context.Context, mods ...util.ModR[Configuration]) (pbEventsV1.EventsV1Client, TestRemoteStore[*pbEventsV1.Event]) {
	start, store := Server(t, ctx, mods...)

	client := tgrpc.NewGRPCClient(t, ctx, pbEventsV1.NewEventsV1Client, start.Address())

	return client, store
}

func Test_Service(t *testing.T) {
	ctx, c := context.WithCancel(context.Background())
	defer c()

	Client(t, ctx)
}
