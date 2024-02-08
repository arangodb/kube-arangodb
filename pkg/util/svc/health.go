//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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

package svc

import (
	"google.golang.org/grpc"
	imHealth "google.golang.org/grpc/health"
	pbHealth "google.golang.org/grpc/health/grpc_health_v1"
)

type HealthState int

const (
	Unhealthy HealthState = iota
	Degraded
	Healthy
)

type HealthType int

const (
	Readiness HealthType = iota
	Liveness
	Startup
)

type Health interface {
	Update(key string, state HealthState)
}

type HealthService interface {
	Service

	Health
}

type emptyHealth struct {
}

func (e emptyHealth) Update(key string, state HealthState) {

}

type health struct {
	*service

	t HealthType
	*imHealth.Server
}

func (h *health) Update(key string, state HealthState) {
	healthState := pbHealth.HealthCheckResponse_UNKNOWN

	switch h.t {
	case Liveness:
		switch state {
		case Healthy, Degraded:
			healthState = pbHealth.HealthCheckResponse_SERVING
		case Unhealthy:
			healthState = pbHealth.HealthCheckResponse_NOT_SERVING
		}
	case Startup, Readiness:
		switch state {
		case Healthy:
			healthState = pbHealth.HealthCheckResponse_SERVING
		case Degraded, Unhealthy:
			healthState = pbHealth.HealthCheckResponse_NOT_SERVING
		}
	}

	h.SetServingStatus(key, healthState)
	h.SetServingStatus("", pbHealth.HealthCheckResponse_SERVING)
}

func (h *health) Name() string {
	return "health"
}

func (h *health) Health() HealthState {
	return Healthy
}

func (h *health) Register(registrar *grpc.Server) {
	pbHealth.RegisterHealthServer(registrar, h)
}

func NewHealthService(cfg Configuration, t HealthType) HealthService {
	health := &health{
		Server: imHealth.NewServer(),
		t:      t,
	}

	svc, err := newService(cfg, health)
	if err != nil {
		return serviceError{err}
	}

	health.service = svc

	return health
}
