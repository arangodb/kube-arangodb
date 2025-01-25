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
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/client-go/tools/clientcmd"

	pbSchedulerV2 "github.com/arangodb/kube-arangodb/integrations/scheduler/v2/definition"
	"github.com/arangodb/kube-arangodb/pkg/logging"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/helm"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
	"github.com/arangodb/kube-arangodb/pkg/util/tests"
	"github.com/arangodb/kube-arangodb/pkg/util/tests/tgrpc"
)

func init() {
	logging.Global().ApplyLogLevels(map[string]logging.Level{
		logging.TopicAll: logging.Debug,
	})
}

func Handler(t *testing.T, ctx context.Context, kclient kclient.Client, client helm.Client, mods ...Mod) svc.Handler {
	handler, err := New(kclient, client, NewConfiguration().With(mods...))
	require.NoError(t, err)

	return handler
}

func InternalClient(t *testing.T, ctx context.Context, mods ...Mod) (pbSchedulerV2.SchedulerV2Client, kclient.Client, helm.Client) {
	client := kclient.NewFakeClient()

	h, err := helm.NewClient(helm.Configuration{
		Namespace: tests.FakeNamespace,
		Config:    client.Config(),
	})
	require.NoError(t, err)

	local := svc.NewService(svc.Configuration{
		Address: "127.0.0.1:0",
	}, Handler(t, ctx, client, h, mods...))

	start := local.Start(ctx)

	return tgrpc.NewGRPCClient(t, ctx, pbSchedulerV2.NewSchedulerV2Client, start.Address()), client, h
}

func ExternalClient(t *testing.T, ctx context.Context, mods ...Mod) (pbSchedulerV2.SchedulerV2Client, kclient.Client, helm.Client) {
	z, ok := os.LookupEnv("TEST_KUBECONFIG")
	if !ok {
		t.Skipf("TEST_KUBECONFIG is not set")
	}

	cfg, err := clientcmd.BuildConfigFromFlags("", z)
	require.NoError(t, err)

	client, err := kclient.NewClient("test", cfg)
	require.NoError(t, err)

	h, err := helm.NewClient(helm.Configuration{
		Namespace: tests.FakeNamespace,
		Config:    client.Config(),
	})
	require.NoError(t, err)

	local := svc.NewService(svc.Configuration{
		Address: "127.0.0.1:0",
	}, Handler(t, ctx, client, h, mods...))

	start := local.Start(ctx)

	return tgrpc.NewGRPCClient(t, ctx, pbSchedulerV2.NewSchedulerV2Client, start.Address()), client, h
}
