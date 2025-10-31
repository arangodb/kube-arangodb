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

	pbSchedulerV2 "github.com/arangodb/kube-arangodb/integrations/scheduler/v2/definition"
	"github.com/arangodb/kube-arangodb/pkg/handlers/platform/chart"
	"github.com/arangodb/kube-arangodb/pkg/logging"
	operator "github.com/arangodb/kube-arangodb/pkg/operatorV2"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/event"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/helm"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient/external"
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

func Client(t *testing.T, ctx context.Context, mods ...Mod) (pbSchedulerV2.SchedulerV2Client, string, kclient.Client, helm.Client) {
	client, ns := external.ExternalClient(t)

	h, err := helm.NewClient(helm.Configuration{
		Namespace: ns,
		Config:    client.Config(),
	})
	require.NoError(t, err)

	mods = append(mods, func(c Configuration) Configuration {
		c.Namespace = ns
		return c
	})

	local, err := svc.NewService(svc.Configuration{
		Address: "127.0.0.1:0",
	}, Handler(t, ctx, client, h, mods...))
	require.NoError(t, err)

	start := local.Start(ctx)

	return tgrpc.NewGRPCClient(t, ctx, pbSchedulerV2.NewSchedulerV2Client, start.Address()), ns, client, h
}

func MockClient(t *testing.T, ctx context.Context, mods ...Mod) (pbSchedulerV2.SchedulerV2Client, string, kclient.Client, helm.Client) {
	client := kclient.NewFakeClient()

	h, err := helm.NewClient(helm.Configuration{
		Namespace: tests.FakeNamespace,
		Config:    client.Config(),
	})
	require.NoError(t, err)

	mods = append(mods, func(c Configuration) Configuration {
		c.Namespace = tests.FakeNamespace
		return c
	})

	local, err := svc.NewService(svc.Configuration{
		Address: "127.0.0.1:0",
	}, Handler(t, ctx, client, h, mods...))
	require.NoError(t, err)

	start := local.Start(ctx)

	return tgrpc.NewGRPCClient(t, ctx, pbSchedulerV2.NewSchedulerV2Client, start.Address()), tests.FakeNamespace, client, h
}

func chartHandler(client kclient.Client, ns string) operator.Handler {
	op := operator.NewOperator("mock", ns, util.Image{Image: "mock"})
	recorder := event.NewEventRecorder("mock", client.Kubernetes())

	return chart.Handler(op, recorder, client)
}
