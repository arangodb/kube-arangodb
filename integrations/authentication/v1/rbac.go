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

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pbAuthenticationV1 "github.com/arangodb/kube-arangodb/integrations/authentication/v1/definition"
	pbImplAuthorizationV1Shared "github.com/arangodb/kube-arangodb/integrations/authorization/v1/shared"
	utilConstantsContext "github.com/arangodb/kube-arangodb/pkg/util/constants/context"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/svc/authenticator"
)

const (
	// ActionCreateToken is the IAM action evaluated before minting a token.
	ActionCreateToken = "authentication:CreateToken"
	// ActionCreateTokenGroup is the IAM action evaluated for each group assigned to a minted token.
	ActionCreateTokenGroup = "authentication:CreateTokenGroup"
)

// rbac wraps a parent AuthenticationV1 integration and enforces RBAC using the authorization evaluator.
// The parent Interface is embedded, so every method is promoted from it by default; rbac overrides only
// the methods that need an authorization check.
type rbac struct {
	Interface

	authz pbImplAuthorizationV1Shared.Evaluator
}

var _ Interface = &rbac{}

// WithRBAC wraps the given AuthenticationV1 integration with RBAC enforcement backed by the authorization
// evaluator provided in the context (AuthZClientPlugin).
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
	pbAuthenticationV1.RegisterAuthenticationV1Server(registrar, r)
}

// Gateway registers the RBAC wrapper (not the wrapped parent) as the HTTP gateway handler, so HTTP calls
// dispatch to the wrapper's authorization-checked overrides.
func (r *rbac) Gateway(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return pbAuthenticationV1.RegisterAuthenticationV1HandlerServer(ctx, mux, r)
}

// CreateToken authorizes the mint before delegating to the wrapped parent. The authorization subject is
// the authenticated caller (the verified identity the server interceptor placed in the context), never
// the request: request.User/Groups are caller-controlled, so using them as the subject would let a caller
// authorize itself by naming a privileged user or claiming roles. The mint is authorized with action
// authentication:CreateToken (resource = the requested user); then every requested group is authorized
// with authentication:CreateTokenGroup (resource = group) so a permitted mint cannot embed groups the
// caller is not allowed to assign. A caller with no identity is rejected as Unauthenticated; a superuser
// (a nil-user server token) is allowed by the SuperUser wrapper.
func (r *rbac) CreateToken(ctx context.Context, request *pbAuthenticationV1.CreateTokenRequest) (*pbAuthenticationV1.CreateTokenResponse, error) {
	caller := authenticator.GetIdentity(ctx)
	if caller == nil {
		return nil, status.Error(codes.Unauthenticated, "Unauthenticated")
	}

	if err := caller.EvaluatePermission(ctx, r.authz, ActionCreateToken, request.GetUser()); err != nil {
		return nil, err
	}

	for _, group := range request.GetGroups() {
		if err := caller.EvaluatePermission(ctx, r.authz, ActionCreateTokenGroup, group); err != nil {
			return nil, err
		}
	}

	return r.Interface.CreateToken(ctx, request)
}
