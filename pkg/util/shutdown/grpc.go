//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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
)

func RegisterCentral(pb grpc.ServiceRegistrar) {
	pbShutdown.RegisterShutdownServer(pb, NewShutdownableShutdownCentralServer())
}

func Register(pb grpc.ServiceRegistrar, closer context.CancelFunc) {
	pbShutdown.RegisterShutdownServer(pb, NewShutdownableShutdownServer(closer))
}

func NewShutdownableShutdownCentralServer() ShutdownableShutdownServer {
	return NewShutdownableShutdownServer(stop)
}

func NewShutdownableShutdownServer(closer context.CancelFunc) ShutdownableShutdownServer {
	return &impl{closer: closer}
}

type ShutdownableShutdownServer interface {
	pbShutdown.ShutdownServer

	Shutdown(cancelFunc context.CancelFunc)
}

var _ ShutdownableShutdownServer = &impl{}

type impl struct {
	pbShutdown.UnimplementedShutdownServer

	closer context.CancelFunc
}

func (i *impl) ShutdownServer(ctx context.Context, empty *server.Empty) (*server.Empty, error) {
	go func() {
		defer i.closer()

		time.Sleep(50 * time.Millisecond)
	}()

	return &server.Empty{}, nil
}

func (i *impl) Shutdown(cancelFunc context.CancelFunc) {
	cancelFunc()
}
