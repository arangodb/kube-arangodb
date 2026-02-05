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

package authorization

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sidecarSvcAuthz "github.com/arangodb/kube-arangodb/pkg/sidecar/services/authorization/definition"
	"github.com/arangodb/kube-arangodb/pkg/sidecar/services/authorization/pool"
	"github.com/arangodb/kube-arangodb/pkg/util/svc/authenticator"
)

func (a *implementation) APIGetPolicy(ctx context.Context, request *sidecarSvcAuthz.AuthorizationAPINamedRequest) (*sidecarSvcAuthz.AuthorizationAPIPolicyResponse, error) {
	if identity := authenticator.GetIdentity(ctx); identity == nil {
		return nil, status.Error(codes.Unauthenticated, "Unauthenticated")
	} else {
		// Restrict only for superuser for now
		if identity.User != nil {
			return nil, status.Error(codes.Unauthenticated, "Only super-user allowed")
		}
	}

	policy, index, ok := a.policies.Item(request.GetName())
	if !ok {
		return nil, status.Error(codes.NotFound, "Policy not found")
	}

	return &sidecarSvcAuthz.AuthorizationAPIPolicyResponse{
		Name:  request.GetName(),
		Index: index,
		Item:  policy,
	}, nil
}

func (a *implementation) APIDeletePolicy(ctx context.Context, request *sidecarSvcAuthz.AuthorizationAPINamedRequest) (*sidecarSvcAuthz.AuthorizationAPIPolicyResponse, error) {
	if identity := authenticator.GetIdentity(ctx); identity == nil {
		return nil, status.Error(codes.Unauthenticated, "Unauthenticated")
	} else {
		// Restrict only for superuser for now
		if identity.User != nil {
			return nil, status.Error(codes.Unauthenticated, "Only super-user allowed")
		}
	}

	offset, err := a.policies.Delete(ctx, request.GetName())

	if err != nil {
		if pool.IsPoolNotFound(err) {
			return nil, status.Error(codes.NotFound, "Policy not found")
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	return &sidecarSvcAuthz.AuthorizationAPIPolicyResponse{
		Name:  request.GetName(),
		Index: offset,
	}, nil
}

func (a *implementation) APICreatePolicy(ctx context.Context, request *sidecarSvcAuthz.AuthorizationAPIPolicyRequest) (*sidecarSvcAuthz.AuthorizationAPIPolicyResponse, error) {
	if identity := authenticator.GetIdentity(ctx); identity == nil {
		return nil, status.Error(codes.Unauthenticated, "Unauthenticated")
	} else {
		// Restrict only for superuser for now
		if identity.User != nil {
			return nil, status.Error(codes.Unauthenticated, "Only super-user allowed")
		}
	}

	if item := request.GetItem(); item == nil {
		return nil, status.Error(codes.InvalidArgument, "Item cannot be empty")
	}

	res, offset, err := a.policies.Create(ctx, request.GetName(), request.GetItem())

	if err != nil {
		if pool.IsPoolAlreadyExistsError(err) {
			return nil, status.Error(codes.AlreadyExists, "Policy already exists")
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	return &sidecarSvcAuthz.AuthorizationAPIPolicyResponse{
		Name:  request.GetName(),
		Index: offset,
		Item:  res,
	}, nil
}

func (a *implementation) APIUpdatePolicy(ctx context.Context, request *sidecarSvcAuthz.AuthorizationAPIPolicyRequest) (*sidecarSvcAuthz.AuthorizationAPIPolicyResponse, error) {
	if identity := authenticator.GetIdentity(ctx); identity == nil {
		return nil, status.Error(codes.Unauthenticated, "Unauthenticated")
	} else {
		// Restrict only for superuser for now
		if identity.User != nil {
			return nil, status.Error(codes.Unauthenticated, "Only super-user allowed")
		}
	}

	if item := request.GetItem(); item == nil {
		return nil, status.Error(codes.InvalidArgument, "Item cannot be empty")
	}

	res, offset, err := a.policies.Update(ctx, request.GetName(), request.GetItem())

	if err != nil {
		if pool.IsPoolNotFound(err) {
			return nil, status.Error(codes.NotFound, "Policy not found")
		}

		if pool.IsPoolNoChangeError(err) {
			return nil, status.Error(codes.AlreadyExists, "Policy exists")
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	return &sidecarSvcAuthz.AuthorizationAPIPolicyResponse{
		Name:  request.GetName(),
		Index: offset,
		Item:  res,
	}, nil
}
