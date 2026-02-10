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
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pbSharedV1 "github.com/arangodb/kube-arangodb/integrations/shared/v1/definition"
	"github.com/arangodb/kube-arangodb/pkg/authorization/service"
	"github.com/arangodb/kube-arangodb/pkg/authorization/types"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod/db"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
)

func NewAuthorizer(client db.Database) svc.Handler {
	return &implementation{
		policies: NewPooler[*types.Policy](client.
			CreateCollection("_policies", db.SourceCollectionProps("_users")).
			WithUniqueIndex("policies_unique_sequence_index", "sequence").
			WithTTLIndex("policies_deleted_index", 30*24*time.Hour, "deleted").
			Get()),
		roles: NewPooler[*types.Role](client.
			CreateCollection("_roles", db.SourceCollectionProps("_users")).
			WithTTLIndex("roles_deleted_index", 30*24*time.Hour, "deleted").
			Get()),
	}
}

var _ service.AuthorizationPoolServiceServer = &implementation{}

type implementation struct {
	service.UnimplementedAuthorizationPoolServiceServer

	policies Pooler[*types.Policy]
	roles    Pooler[*types.Role]
}

func (a *implementation) Name() string {
	return "implementation"
}

func (a *implementation) Health(ctx context.Context) svc.HealthState {
	return svc.Healthy
}

func (a *implementation) Register(registrar *grpc.Server) {
	service.RegisterAuthorizationPoolServiceServer(registrar, a)
}

func (a *implementation) Gateway(ctx context.Context, mux *runtime.ServeMux) error {
	return nil
}

func (a *implementation) PoolPolicyChanges(request *service.AuthorizationPoolRequest, g grpc.ServerStreamingServer[service.AuthorizationPoolPolicyResponse]) error {
	index := request.GetStart()

	tickerT := time.NewTicker(time.Second)
	defer tickerT.Stop()

	last := time.Now()

	for {
		select {
		case <-tickerT.C:
			// Process
			items, err := a.policies.Pool(index)
			if err != nil {
				var poolOutOfBoundsError PoolOutOfBoundsError
				if errors.As(err, &poolOutOfBoundsError) {
					return status.Error(codes.OutOfRange, "out of bounds")
				}
			}

			if len(items) == 0 && time.Since(last) > request.GetTimeout().AsDuration() {
				// Send empty response
				if err := g.Send(&service.AuthorizationPoolPolicyResponse{
					Items: nil,
				}); err != nil {
					return status.Error(codes.Internal, err.Error())
				}
			} else {
				continue
			}

			for _, item := range util.BatchList(128, items) {
				if err := g.Send(&service.AuthorizationPoolPolicyResponse{
					Items: util.FormatList(item, func(a OffsetItem[*types.Policy]) *service.AuthorizationPoolPolicyResponseItem {
						return &service.AuthorizationPoolPolicyResponseItem{
							Name:  a.Name,
							Index: a.Sequence,
							Item:  a.Item,
						}
					}),
				}); err != nil {
					return status.Error(codes.Internal, err.Error())
				}
			}

			last = time.Now()

		case <-g.Context().Done():
			return g.Context().Err()
		}
	}
}

func (a *implementation) GetPolicy(empty *pbSharedV1.Empty, g grpc.ServerStreamingServer[service.AuthorizationPoolPolicyResponse]) error {
	items := a.policies.Get()

	for _, item := range util.BatchList(128, items) {
		if err := g.Send(&service.AuthorizationPoolPolicyResponse{
			Items: util.FormatList(item, func(a OffsetItem[*types.Policy]) *service.AuthorizationPoolPolicyResponseItem {
				return &service.AuthorizationPoolPolicyResponseItem{
					Name:  a.Name,
					Index: a.Sequence,
					Item:  a.Item,
				}
			}),
		}); err != nil {
			return err
		}
	}

	return nil
}

func (a *implementation) PoolRoleChanges(request *service.AuthorizationPoolRequest, g grpc.ServerStreamingServer[service.AuthorizationPoolRoleResponse]) error {
	index := request.GetStart()

	tickerT := time.NewTicker(time.Second)
	defer tickerT.Stop()

	last := time.Now()

	for {
		select {
		case <-tickerT.C:
			// Process
			items, err := a.roles.Pool(index)
			if err != nil {
				var poolOutOfBoundsError PoolOutOfBoundsError
				if errors.As(err, &poolOutOfBoundsError) {
					return status.Error(codes.OutOfRange, "out of bounds")
				}
			}

			if len(items) == 0 && time.Since(last) > request.GetTimeout().AsDuration() {
				// Send empty response
				if err := g.Send(&service.AuthorizationPoolRoleResponse{
					Items: nil,
				}); err != nil {
					return status.Error(codes.Internal, err.Error())
				}
			} else {
				continue
			}

			for _, item := range util.BatchList(128, items) {
				if err := g.Send(&service.AuthorizationPoolRoleResponse{
					Items: util.FormatList(item, func(a OffsetItem[*types.Role]) *service.AuthorizationPoolRoleResponseItem {
						return &service.AuthorizationPoolRoleResponseItem{
							Name:  a.Name,
							Index: a.Sequence,
							Item:  a.Item,
						}
					}),
				}); err != nil {
					return status.Error(codes.Internal, err.Error())
				}
			}

			last = time.Now()

		case <-g.Context().Done():
			return g.Context().Err()
		}
	}
}

func (a *implementation) GetRole(empty *pbSharedV1.Empty, g grpc.ServerStreamingServer[service.AuthorizationPoolRoleResponse]) error {
	items := a.roles.Get()

	for _, item := range util.BatchList(128, items) {
		if err := g.Send(&service.AuthorizationPoolRoleResponse{
			Items: util.FormatList(item, func(a OffsetItem[*types.Role]) *service.AuthorizationPoolRoleResponseItem {
				return &service.AuthorizationPoolRoleResponseItem{
					Name:  a.Name,
					Index: a.Sequence,
					Item:  a.Item,
				}
			}),
		}); err != nil {
			return err
		}
	}

	return nil
}

func (a *implementation) Refresh(ctx context.Context) error {
	if err := a.policies.Refresh(ctx); err != nil {
		return err
	}
	if err := a.roles.Refresh(ctx); err != nil {
		return err
	}

	return nil
}
