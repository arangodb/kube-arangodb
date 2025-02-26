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

package v0

import (
	"context"
	"fmt"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"

	pbAuthorizationV0 "github.com/arangodb/kube-arangodb/integrations/authorization/v0/definition"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
)

var _ pbAuthorizationV0.AuthorizationV0Server = &implementation{}
var _ svc.Handler = &implementation{}

func New() svc.Handler {
	return newInternal()
}

func newInternal() *implementation {
	return &implementation{}
}

type implementation struct {
	pbAuthorizationV0.UnimplementedAuthorizationV0Server
}

func (i *implementation) Name() string {
	return pbAuthorizationV0.Name
}

func (i *implementation) Health() svc.HealthState {
	return svc.Healthy
}

func (i *implementation) Register(registrar *grpc.Server) {
	pbAuthorizationV0.RegisterAuthorizationV0Server(registrar, i)
}

func (i *implementation) Gateway(ctx context.Context, mux *runtime.ServeMux) error {
	return nil
}

func (i *implementation) Can(ctx context.Context, request *pbAuthorizationV0.CanRequest) (*pbAuthorizationV0.CanResponse, error) {
	if request == nil {
		return nil, errors.Errorf("Request is nil")
	}

	return &pbAuthorizationV0.CanResponse{
		Allowed: true,
		Message: fmt.Sprintf("Access by user `%s` to resource `%s/%s` has been granted", request.GetUser(), request.GetApi(), request.GetAction()),
	}, nil
}
