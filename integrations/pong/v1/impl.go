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
	"google.golang.org/protobuf/types/known/timestamppb"

	pbPongV1 "github.com/arangodb/kube-arangodb/integrations/pong/v1/definition"
	pbSharedV1 "github.com/arangodb/kube-arangodb/integrations/shared/v1/definition"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
)

type Service struct {
	Name, Version string
	Enabled       bool
}

func (s Service) asService() *pbPongV1.PongV1Service {
	return &pbPongV1.PongV1Service{
		Name:    s.Name,
		Version: s.Version,
		Enabled: s.Enabled,
	}
}

func New(services ...Service) (svc.Handler, error) {
	pServices := make(util.List[Service], 0, len(services))

	for _, svc := range services {
		if pServices.Contains(func(service Service) bool {
			return service.Name == svc.Name && service.Version == svc.Version
		}) {
			return nil, errors.Errorf("Service %s/%s already registered", svc.Name, svc.Version)
		}

		pServices = pServices.Append(svc)
	}

	pServices = pServices.Sort(func(a Service, b Service) bool {
		if a.Name == b.Name {
			return a.Version < b.Version
		}

		return a.Name < b.Name
	})

	return &impl{services: pServices.List()}, nil
}

var _ pbPongV1.PongV1Server = &impl{}
var _ svc.Handler = &impl{}

type impl struct {
	services []Service

	pbPongV1.UnimplementedPongV1Server
}

func (i *impl) Name() string {
	return pbPongV1.Name
}

func (i *impl) Health() svc.HealthState {
	return svc.Healthy
}

func (i *impl) Register(registrar *grpc.Server) {
	pbPongV1.RegisterPongV1Server(registrar, i)
}

func (i *impl) Ping(context.Context, *pbSharedV1.Empty) (*pbPongV1.PongV1PingResponse, error) {
	return &pbPongV1.PongV1PingResponse{Time: timestamppb.New(time.Now().UTC())}, nil
}

func (i *impl) Services(context.Context, *pbSharedV1.Empty) (*pbPongV1.PongV1ServicesResponse, error) {
	var r = make([]*pbPongV1.PongV1Service, len(i.services))

	for id := range i.services {
		r[id] = i.services[id].asService()
	}

	return &pbPongV1.PongV1ServicesResponse{
		Services: r,
	}, nil
}
