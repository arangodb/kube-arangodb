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

	pbSharedV1 "github.com/arangodb/kube-arangodb/integrations/shared/v1/definition"
	sidecarSvcAuthzDefinition "github.com/arangodb/kube-arangodb/pkg/sidecar/services/authorization/definition"
	"github.com/arangodb/kube-arangodb/pkg/sidecar/services/authorization/pool"
	sidecarSvcAuthzTypes "github.com/arangodb/kube-arangodb/pkg/sidecar/services/authorization/types"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/svc/authenticator"
)

func (a *implementation) APIListPolicy(ctx context.Context, empty *pbSharedV1.Empty) (*sidecarSvcAuthzDefinition.AuthorizationAPIListResponse, error) {
	if err := a.Health(ctx).Require(); err != nil {
		return nil, err
	}

	if err := authenticator.GetIdentity(ctx).EvaluatePermission(ctx, a.auth, "rbac:ListPolicy", ""); err != nil {
		return nil, err
	}

	return &sidecarSvcAuthzDefinition.AuthorizationAPIListResponse{
		Names: util.FormatList(a.policies.Get(), func(a pool.OffsetItem[*sidecarSvcAuthzTypes.Policy]) string {
			return a.Name
		}),
	}, nil
}

func (a *implementation) APIGetPolicy(ctx context.Context, request *sidecarSvcAuthzDefinition.AuthorizationAPINamedRequest) (*sidecarSvcAuthzDefinition.AuthorizationAPIPolicyResponse, error) {
	if err := a.Health(ctx).Require(); err != nil {
		return nil, err
	}

	if err := authenticator.GetIdentity(ctx).EvaluatePermission(ctx, a.auth, "rbac:GetPolicy", request.GetName()); err != nil {
		return nil, err
	}

	policy, index, ok := a.policies.Item(request.GetName())
	if !ok {
		return nil, status.Error(codes.NotFound, "Policy not found")
	}

	return &sidecarSvcAuthzDefinition.AuthorizationAPIPolicyResponse{
		Name:  request.GetName(),
		Index: index,
		Item:  policy,
	}, nil
}

func (a *implementation) APIDeletePolicy(ctx context.Context, request *sidecarSvcAuthzDefinition.AuthorizationAPINamedRequest) (*sidecarSvcAuthzDefinition.AuthorizationAPIPolicyResponse, error) {
	if err := a.Health(ctx).Require(); err != nil {
		return nil, err
	}

	if err := authenticator.GetIdentity(ctx).EvaluatePermission(ctx, a.auth, "rbac:DeletePolicy", request.GetName()); err != nil {
		return nil, err
	}

	offset, err := a.policies.Delete(ctx, request.GetName())

	if err != nil {
		if pool.IsPoolNotFound(err) {
			return nil, status.Error(codes.NotFound, "Policy not found")
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	return &sidecarSvcAuthzDefinition.AuthorizationAPIPolicyResponse{
		Name:  request.GetName(),
		Index: offset,
	}, nil
}

func (a *implementation) APICreatePolicy(ctx context.Context, request *sidecarSvcAuthzDefinition.AuthorizationAPIPolicyRequest) (*sidecarSvcAuthzDefinition.AuthorizationAPIPolicyResponse, error) {
	if err := a.Health(ctx).Require(); err != nil {
		return nil, err
	}

	if err := authenticator.GetIdentity(ctx).EvaluatePermission(ctx, a.auth, "rbac:CreatePolicy", request.GetName()); err != nil {
		return nil, err
	}

	if item := request.GetName(); item == "" {
		return nil, status.Error(codes.InvalidArgument, "Name cannot be empty")
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

	return &sidecarSvcAuthzDefinition.AuthorizationAPIPolicyResponse{
		Name:  request.GetName(),
		Index: offset,
		Item:  res,
	}, nil
}

func (a *implementation) APIUpdatePolicy(ctx context.Context, request *sidecarSvcAuthzDefinition.AuthorizationAPIPolicyRequest) (*sidecarSvcAuthzDefinition.AuthorizationAPIPolicyResponse, error) {
	if err := a.Health(ctx).Require(); err != nil {
		return nil, err
	}

	if err := authenticator.GetIdentity(ctx).EvaluatePermission(ctx, a.auth, "rbac:UpdatePolicy", request.GetName()); err != nil {
		return nil, err
	}

	if item := request.GetName(); item == "" {
		return nil, status.Error(codes.InvalidArgument, "Name cannot be empty")
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

	return &sidecarSvcAuthzDefinition.AuthorizationAPIPolicyResponse{
		Name:  request.GetName(),
		Index: offset,
		Item:  res,
	}, nil
}
