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

package probe

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	pbHealth "google.golang.org/grpc/health/grpc_health_v1"
)

type HealthService interface {
	Register(server *grpc.Server)
	SetServing()
	Shutdown()
}

func NewHealthService() HealthService {
	return &grpcHealthService{
		hs: health.NewServer(),
	}
}

type grpcHealthService struct {
	hs *health.Server
}

func (s *grpcHealthService) Register(server *grpc.Server) {
	pbHealth.RegisterHealthServer(server, s.hs)
}

// SetServing marks the health response as Serving for all services
func (s *grpcHealthService) SetServing() {
	s.hs.SetServingStatus("", pbHealth.HealthCheckResponse_SERVING)
}

// Shutdown marks as not serving and forbids further changes in health status
func (s *grpcHealthService) Shutdown() {
	s.hs.Shutdown()
}
