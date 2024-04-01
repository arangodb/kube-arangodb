//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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

	pbSchedulerV1 "github.com/arangodb/kube-arangodb/integrations/scheduler/v1/definition"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
	"github.com/arangodb/kube-arangodb/pkg/util/tests/tgrpc"
)

func Handler(t *testing.T, ctx context.Context, client kclient.Client, mods ...Mod) svc.Handler {
	handler, err := New(ctx, client, NewConfiguration().With(mods...))
	require.NoError(t, err)

	return handler
}

func Client(t *testing.T, ctx context.Context, client kclient.Client, mods ...Mod) pbSchedulerV1.SchedulerV1Client {
	local := svc.NewService(svc.Configuration{
		Address: "127.0.0.1:0",
	}, Handler(t, ctx, client, mods...))

	start := local.Start(ctx)

	return tgrpc.NewGRPCClient(t, ctx, pbSchedulerV1.NewSchedulerV1Client, start.Address())
}
