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

	"google.golang.org/grpc"

	pbSchedulerV2 "github.com/arangodb/kube-arangodb/integrations/scheduler/v2/definition"
	pbSharedV1 "github.com/arangodb/kube-arangodb/integrations/shared/v1/definition"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/helm"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
)

var _ pbSchedulerV2.SchedulerV2Server = &implementation{}
var _ svc.Handler = &implementation{}

func New(kclient kclient.Client, client helm.Client, cfg Configuration) (svc.Handler, error) {
	return newInternal(kclient, client, cfg)
}

func newInternal(kclient kclient.Client, client helm.Client, c Configuration) (*implementation, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}

	return &implementation{
		cfg:     c,
		client:  client,
		kclient: kclient,
	}, nil
}

type implementation struct {
	cfg Configuration

	kclient kclient.Client
	client  helm.Client

	pbSchedulerV2.UnimplementedSchedulerV2Server
}

func (i *implementation) Name() string {
	return pbSchedulerV2.Name
}

func (i *implementation) Register(registrar *grpc.Server) {
	pbSchedulerV2.RegisterSchedulerV2Server(registrar, i)
}

func (i *implementation) Health() svc.HealthState {
	return svc.Healthy
}

func (i *implementation) InvalidateCache(ctx context.Context, in *pbSharedV1.Empty) (*pbSharedV1.Empty, error) {
	i.client.Invalidate()

	return &pbSharedV1.Empty{}, nil
}
