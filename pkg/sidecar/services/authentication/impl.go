//
// DISCLAIMER
//
// Copyright 2026 ArangoDB GmbH, Cologne, Germany
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

package authentication

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pbSharedV1 "github.com/arangodb/kube-arangodb/integrations/shared/v1/definition"
	sidecarSvcAuthnDefinition "github.com/arangodb/kube-arangodb/pkg/sidecar/services/authentication/definition"
	"github.com/arangodb/kube-arangodb/pkg/util/cache"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
	"github.com/arangodb/kube-arangodb/pkg/util/svc/authenticator"
	utilToken "github.com/arangodb/kube-arangodb/pkg/util/token"
)

func NewWithSecret(secret cache.Object[utilToken.Secret]) svc.HandlerGateway {
	return New(secretSigningKeys{secret: secret})
}

func NewWithEmpty() svc.HandlerGateway {
	return New(emptySigningKeys{})
}

func New(keys SigningKeys) svc.HandlerGateway {
	return &implementation{keys: keys}
}

type implementation struct {
	sidecarSvcAuthnDefinition.UnimplementedSidecarAuthenticationServiceServer

	keys SigningKeys
}

func (i implementation) GetKeys(ctx context.Context, empty *pbSharedV1.Empty) (*sidecarSvcAuthnDefinition.SidecarAuthenticationKeysResponse, error) {
	return i.GetOptionalKeys(ctx, &sidecarSvcAuthnDefinition.SidecarAuthenticationKeysRequest{})
}

func (i implementation) GetOptionalKeys(ctx context.Context, request *sidecarSvcAuthnDefinition.SidecarAuthenticationKeysRequest) (*sidecarSvcAuthnDefinition.SidecarAuthenticationKeysResponse, error) {
	if auth := authenticator.GetIdentity(ctx); auth == nil {
		return nil, status.Error(codes.Unauthenticated, "Unauthorized")
	}

	if !i.keys.Enabled() {
		return nil, status.Error(codes.Unavailable, "Keys not enabled")
	}

	keys, hash, err := i.keys.Get(ctx)
	if err != nil {
		logger.Err(err).Warn("Failed to get secret")
		return nil, status.Error(codes.Internal, err.Error())
	}

	if z := request.Checksum; z != nil {
		if hash == *z {
			return nil, status.Error(codes.AlreadyExists, "Secret hash did not change")
		}
	}

	return &sidecarSvcAuthnDefinition.SidecarAuthenticationKeysResponse{
		Keys:     keys,
		Checksum: hash,
	}, nil
}

func (i implementation) Name() string {
	return "authentication"
}

func (i implementation) Health(ctx context.Context) svc.HealthState {
	return svc.Healthy
}

func (i implementation) Register(registrar *grpc.Server) {
	sidecarSvcAuthnDefinition.RegisterSidecarAuthenticationServiceServer(registrar, i)
}

func (i implementation) Gateway(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return sidecarSvcAuthnDefinition.RegisterSidecarAuthenticationServiceHandler(ctx, mux, conn)
}
