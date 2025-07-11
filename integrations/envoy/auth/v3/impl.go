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

package v3

import (
	"context"
	"time"

	pbEnvoyAuthV3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"k8s.io/apimachinery/pkg/util/uuid"

	impl2 "github.com/arangodb/kube-arangodb/integrations/envoy/auth/v3/impl"
	pbImplEnvoyAuthV3Shared "github.com/arangodb/kube-arangodb/integrations/envoy/auth/v3/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/errors/panics"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
)

func New(ctx context.Context, config pbImplEnvoyAuthV3Shared.Configuration) svc.Handler {
	return &impl{
		config:  config,
		handler: impl2.Factory().Render(ctx, config),
	}
}

var _ pbEnvoyAuthV3.AuthorizationServer = &impl{}
var _ svc.Handler = &impl{}

type impl struct {
	pbEnvoyAuthV3.UnimplementedAuthorizationServer

	config pbImplEnvoyAuthV3Shared.Configuration

	handler pbImplEnvoyAuthV3Shared.AuthHandler
}

func (i *impl) Name() string {
	return pbImplEnvoyAuthV3Shared.Name
}

func (i *impl) Health() svc.HealthState {
	return svc.Healthy
}

func (i *impl) Register(registrar *grpc.Server) {
	pbEnvoyAuthV3.RegisterAuthorizationServer(registrar, i)
}

func (i *impl) Gateway(ctx context.Context, mux *runtime.ServeMux) error {
	return nil
}

func (i *impl) Check(ctx context.Context, request *pbEnvoyAuthV3.CheckRequest) (resp *pbEnvoyAuthV3.CheckResponse, err error) {
	q := logger

	q = q.Str("id", string(uuid.NewUUID()))

	start := time.Now()

	if request != nil {
		if attr := request.Attributes; attr != nil {
			if req := attr.Request; req != nil {
				if http := req.Http; http != nil {
					q = q.Str("path", http.Path).Str("method", http.Method)
				}
			}
		}
	}

	defer func() {
		if resp != nil {
			if status := resp.Status; status != nil {
				q = q.Int32("response_code", status.Code)
			}
		}

		if err != nil {
			q = q.Err(err)
			q.Dur("duration", time.Since(start)).Warn("Request Completed with error")
		} else {
			q.Dur("duration", time.Since(start)).Debug("Request Completed")
		}
	}()

	q.Debug("Request Started")

	resp, err = panics.RecoverO1(func() (*pbEnvoyAuthV3.CheckResponse, error) {
		return i.check(ctx, request)
	})

	if err != nil {
		var v pbImplEnvoyAuthV3Shared.CustomResponse
		if errors.As(err, &v) {
			return v.Response()
		}
		return nil, err
	}

	return resp, nil
}

func (i *impl) check(ctx context.Context, request *pbEnvoyAuthV3.CheckRequest) (*pbEnvoyAuthV3.CheckResponse, error) {
	var auth pbImplEnvoyAuthV3Shared.Response
	if err := i.handler.Handle(ctx, request, &auth); err != nil {
		return nil, err
	}

	return auth.AsResponse(), nil
}
