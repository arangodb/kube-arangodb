//
// DISCLAIMER
//
// Copyright 2024-2025 ArangoDB GmbH, Cologne, Germany
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

package v2

import (
	"context"
	_ "embed"
	"testing"

	"github.com/stretchr/testify/require"

	pbStorageV2 "github.com/arangodb/kube-arangodb/integrations/storage/v2/definition"
	"github.com/arangodb/kube-arangodb/pkg/logging"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
	"github.com/arangodb/kube-arangodb/pkg/util/tests/tgrpc"
)

func init() {
	logging.Global().ApplyLogLevels(map[string]logging.Level{
		logging.TopicAll: logging.Debug,
	})
}

type configGenerator func(t *testing.T, mods ...util.ModR[Configuration]) Configuration

func Handler(t *testing.T, gen configGenerator, mods ...util.ModR[Configuration]) svc.Handler {
	handler, err := New(gen(t, mods...))
	require.NoError(t, err)

	return handler
}

func Client(t *testing.T, ctx context.Context, gen configGenerator, mods ...util.ModR[Configuration]) pbStorageV2.StorageV2Client {
	local, err := svc.NewService(svc.Configuration{
		Address: "127.0.0.1:0",
	}, Handler(t, gen, mods...))
	require.NoError(t, err)

	start := local.Start(ctx)

	return tgrpc.NewGRPCClient(t, ctx, pbStorageV2.NewStorageV2Client, start.Address())
}
