//
// DISCLAIMER
//
// Copyright 2023-2024 ArangoDB GmbH, Cologne, Germany
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
	"time"

	"google.golang.org/grpc"

	pbSharedV1 "github.com/arangodb/kube-arangodb/integrations/shared/v1/definition"
	pbShutdownV1 "github.com/arangodb/kube-arangodb/integrations/shutdown/v1/definition"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
)

func New(closer context.CancelFunc) svc.Handler {
	return &impl{closer: closer}
}

var _ pbShutdownV1.ShutdownV1Server = &impl{}
var _ svc.Handler = &impl{}

type impl struct {
	pbShutdownV1.UnimplementedShutdownV1Server

	closer context.CancelFunc
}

func (i *impl) Name() string {
	return Name
}

func (i *impl) Health() svc.HealthState {
	return svc.Healthy
}

func (i *impl) Register(registrar *grpc.Server) {
	pbShutdownV1.RegisterShutdownV1Server(registrar, i)
}

func (i *impl) Shutdown(ctx context.Context, empty *pbSharedV1.Empty) (*pbSharedV1.Empty, error) {
	go func() {
		defer i.closer()

		time.Sleep(50 * time.Millisecond)
	}()

	return &pbSharedV1.Empty{}, nil
}
