//
// DISCLAIMER
//
// Copyright 2024-2026 ArangoDB GmbH, Cologne, Germany
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
	"context"
	"sync"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	imHealth "google.golang.org/grpc/health"
	pbHealth "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

type HealthState int

const (
	Unhealthy HealthState = iota
	Degraded
	Healthy
)

func (h HealthState) String() string {
	switch h {
	case Unhealthy:
		return "unhealthy"
	case Degraded:
		return "degraded"
	case Healthy:
		return "healthy"
	default:
		return "unhealthy"
	}
}

func (h HealthState) Require() error {
	if h == Healthy {
		return nil
	}

	return status.Error(codes.Unavailable, "service is not healthy")
}

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
	Handler

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

	health map[string]HealthState
	lock   sync.Mutex
}

func (h *health) Update(key string, state HealthState) {
	h.lock.Lock()
	defer h.lock.Unlock()

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
	if v, ok := h.health[key]; ok && v == state {
		return
	}

	h.health[key] = state

	l := logger
	for k, v := range h.health {
		l = l.Str(k, v.String())
	}

	if v := util.MapValues(h.health); util.ContainsList(v, Unhealthy) || util.ContainsList(v, Degraded) {
		h.SetServingStatus("", pbHealth.HealthCheckResponse_NOT_SERVING)
		l.Warn("Health check unhealthy")
	} else {
		h.SetServingStatus("", pbHealth.HealthCheckResponse_SERVING)
		l.Info("Health check healthy")
	}
}

func (h *health) Name() string {
	return "health"
}

func (h *health) Health(ctx context.Context) HealthState {
	return Healthy
}

func (h *health) Register(registrar *grpc.Server) {
	pbHealth.RegisterHealthServer(registrar, h)
}

func (h *health) Gateway(ctx context.Context, mux *runtime.ServeMux) error {
	return nil
}

func NewHealthService(cfg Configuration, t HealthType, handlers ...Handler) (HealthService, error) {
	health := &health{
		Server: imHealth.NewServer(),
		t:      t,
		health: make(map[string]HealthState),
	}

	var h []Handler
	h = append(h, health)
	h = append(h, handlers...)

	z, err := newService(cfg, h...)
	if err != nil {
		return nil, err
	}

	health.service = z

	return health, nil
}
