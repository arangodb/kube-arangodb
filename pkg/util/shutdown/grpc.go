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

package shutdown

import (
	"context"
	"time"

	"google.golang.org/grpc"

	"github.com/arangodb/kube-arangodb/pkg/api/server"
	pbShutdown "github.com/arangodb/kube-arangodb/pkg/api/shutdown/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
)

func NewGlobalShutdownServer() svc.Handler {
	return NewShutdownServer(stop)
}

func NewShutdownServer(closer context.CancelFunc) svc.Handler {
	return &impl{closer: closer}
}

var _ pbShutdown.ShutdownServer = &impl{}
var _ svc.Handler = &impl{}

type impl struct {
	pbShutdown.UnimplementedShutdownServer

	closer context.CancelFunc
}

func (i *impl) Name() string {
	return "shutdown"
}

func (i *impl) Health() svc.HealthState {
	return svc.Healthy
}

func (i *impl) Register(registrar *grpc.Server) {
	pbShutdown.RegisterShutdownServer(registrar, i)
}

func (i *impl) ShutdownServer(ctx context.Context, empty *server.Empty) (*server.Empty, error) {
	go func() {
		defer i.closer()

		time.Sleep(50 * time.Millisecond)
	}()

	return &server.Empty{}, nil
}
