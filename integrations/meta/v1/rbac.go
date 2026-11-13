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

package v1

import (
	"context"

	"google.golang.org/grpc"

	pbImplAuthorizationV1Shared "github.com/arangodb/kube-arangodb/integrations/authorization/v1/shared"
	pbMetaV1 "github.com/arangodb/kube-arangodb/integrations/meta/v1/definition"
	pbSharedV1 "github.com/arangodb/kube-arangodb/integrations/shared/v1/definition"
	"github.com/arangodb/kube-arangodb/pkg/util"
	utilConstantsContext "github.com/arangodb/kube-arangodb/pkg/util/constants/context"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/svc/authenticator"
)

const (
	// ActionGetKey is the IAM action evaluated before reading a key.
	ActionGetKey = "meta:GetKey"
	// ActionUpdateKey is the IAM action evaluated before creating, updating or moving a key.
	ActionUpdateKey = "meta:UpdateKey"
	// ActionDeleteKey is the IAM action evaluated before deleting a key.
	ActionDeleteKey = "meta:DeleteKey"
	// ActionListKey is the IAM action evaluated before listing keys.
	ActionListKey = "meta:ListKey"
)

// rbac wraps a parent MetaV1 integration and enforces RBAC using the authorization evaluator. The parent
// Interface is embedded, so every method is promoted from it by default; rbac overrides only the methods
// that need an authorization check.
type rbac struct {
	Interface

	authz pbImplAuthorizationV1Shared.Evaluator
}

var _ Interface = &rbac{}

// WithRBAC wraps the given MetaV1 integration with RBAC enforcement backed by the authorization evaluator
// provided in the context (AuthZClientPlugin).
func WithRBAC(ctx context.Context, in Interface) (Interface, error) {
	authz, ok := utilConstantsContext.AuthZClientPlugin.Get(ctx)
	if !ok {
		return nil, errors.Errorf("Unable to get AuthZ Client Plugin")
	}

	return &rbac{
		Interface: in,
		authz:     authz,
	}, nil
}

// Register registers the RBAC wrapper (not the wrapped parent) as the gRPC server, so incoming calls
// dispatch to the wrapper's authorization-checked overrides.
func (r *rbac) Register(registrar *grpc.Server) {
	pbMetaV1.RegisterMetaV1Server(registrar, r)
}

func (r *rbac) Get(ctx context.Context, req *pbMetaV1.ObjectRequest) (*pbMetaV1.ObjectResponse, error) {
	if err := authenticator.GetIdentity(ctx).EvaluatePermission(ctx, r.authz, ActionGetKey, r.Key(req.GetKey())); err != nil {
		return nil, err
	}

	return r.Interface.Get(ctx, req)
}

func (r *rbac) GetBatch(ctx context.Context, request *pbMetaV1.ObjectBatchRequest) (*pbMetaV1.ObjectBatchResponse, error) {
	resp, err := util.ParallelProcessOutputErr(func(in *pbMetaV1.ObjectRequest) (*pbMetaV1.ObjectResponse, error) {
		return r.Get(ctx, in)
	}, 4, request.Items)
	if err != nil {
		return nil, err
	}

	return &pbMetaV1.ObjectBatchResponse{Items: resp}, nil
}

func (r *rbac) Set(ctx context.Context, req *pbMetaV1.SetRequest) (*pbMetaV1.ObjectResponse, error) {
	if err := authenticator.GetIdentity(ctx).EvaluatePermission(ctx, r.authz, ActionUpdateKey, r.Key(req.GetKey())); err != nil {
		return nil, err
	}

	return r.Interface.Set(ctx, req)
}

func (r *rbac) Delete(ctx context.Context, req *pbMetaV1.ObjectRequest) (*pbSharedV1.Empty, error) {
	if err := authenticator.GetIdentity(ctx).EvaluatePermission(ctx, r.authz, ActionDeleteKey, r.Key(req.GetKey())); err != nil {
		return nil, err
	}

	return r.Interface.Delete(ctx, req)
}

func (r *rbac) Move(ctx context.Context, req *pbMetaV1.MoveRequest) (*pbMetaV1.ObjectResponse, error) {
	// Move removes the source object and creates the destination one, so the update permission is required
	// on both keys.
	if err := authenticator.GetIdentity(ctx).EvaluatePermission(ctx, r.authz, ActionUpdateKey, r.Key(req.GetSource())); err != nil {
		return nil, err
	}

	if err := authenticator.GetIdentity(ctx).EvaluatePermission(ctx, r.authz, ActionUpdateKey, r.Key(req.GetDestination())); err != nil {
		return nil, err
	}

	return r.Interface.Move(ctx, req)
}

func (r *rbac) List(req *pbMetaV1.ListRequest, server pbMetaV1.MetaV1_ListServer) error {
	if err := authenticator.GetIdentity(server.Context()).EvaluatePermission(server.Context(), r.authz, ActionListKey, req.GetPrefix()); err != nil {
		return err
	}

	return r.Interface.List(req, server)
}
