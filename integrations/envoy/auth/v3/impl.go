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

package v3

import (
	"context"

	pbEnvoyAuthV3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	"google.golang.org/grpc"

	"github.com/arangodb/kube-arangodb/pkg/util/svc"
)

func New() svc.Handler {
	return &impl{}
}

var _ pbEnvoyAuthV3.AuthorizationServer = &impl{}
var _ svc.Handler = &impl{}

type impl struct {
	pbEnvoyAuthV3.UnimplementedAuthorizationServer
}

func (i *impl) Name() string {
	return Name
}

func (i *impl) Health() svc.HealthState {
	return svc.Healthy
}

func (i *impl) Register(registrar *grpc.Server) {
	pbEnvoyAuthV3.RegisterAuthorizationServer(registrar, i)
}

func (i *impl) Check(ctx context.Context, request *pbEnvoyAuthV3.CheckRequest) (*pbEnvoyAuthV3.CheckResponse, error) {
	return &pbEnvoyAuthV3.CheckResponse{}, nil
}
