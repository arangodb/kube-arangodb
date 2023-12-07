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
	"net"

	"google.golang.org/grpc"

	"github.com/arangodb/kube-arangodb/pkg/util/svc"
)

type ServiceConfig struct {
	ListenAddress string
}

func ServiceCentral(config ServiceConfig) svc.Service {
	server := grpc.NewServer( /* currently no auth parameters required */ )

	RegisterCentral(server)

	return &service{
		cfg:        config,
		grpcServer: server,
	}
}

func Service(config ServiceConfig, closer context.CancelFunc) svc.Service {
	server := grpc.NewServer( /* currently no auth parameters required */ )

	Register(server, closer)

	return &service{
		cfg:        config,
		grpcServer: server,
	}
}

type service struct {
	grpcServer *grpc.Server
	cfg        ServiceConfig
}

func (s *service) Run(ctx context.Context) error {
	ln, err := net.Listen("tcp", s.cfg.ListenAddress)
	if err != nil {
		return err
	}
	defer ln.Close()

	errChan := make(chan error)
	go func() {
		if serveErr := s.grpcServer.Serve(ln); serveErr != nil && serveErr != grpc.ErrServerStopped {
			errChan <- serveErr
		}
	}()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	select {
	case <-ctx.Done():
		s.grpcServer.GracefulStop()
	case err = <-errChan:
		s.grpcServer.Stop()
		close(errChan)
	}

	return err
}
