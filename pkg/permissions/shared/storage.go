//
// DISCLAIMER
//
// Copyright 2025 ArangoDB GmbH, Cologne, Germany
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

package shared

import (
	"context"

	pbAuthorizationV1 "github.com/arangodb/kube-arangodb/integrations/authorization/v1/definition"
)

type Evaluator interface {
	Evaluate(ctx context.Context, request *pbAuthorizationV1.AuthorizationV1PermissionRequest) (*pbAuthorizationV1.AuthorizationV1PermissionResponse, error)
}

type Storage interface {
	Evaluator

	Init(ctx context.Context) error
	Clean(ctx context.Context) error

	GetRole(ctx context.Context, name string) (*pbAuthorizationV1.AuthorizationV1Role, error)
	GetRoles(ctx context.Context) ([]*pbAuthorizationV1.AuthorizationV1Role, error)
	DeleteRole(ctx context.Context, name string) error

	RolePolicies(ctx context.Context, name string) ([]*pbAuthorizationV1.AuthorizationV1Policy, error)
	AssignRoleToPolicy(ctx context.Context, role, policy string) error
	DetachRoleFromPolicy(ctx context.Context, role, policy string) error

	UserPolicies(ctx context.Context, name string) ([]*pbAuthorizationV1.AuthorizationV1Policy, error)
	AssignUserToPolicy(ctx context.Context, user, policy string) error
	DetachUserFromPolicy(ctx context.Context, user, policy string) error

	GetPolicy(ctx context.Context, name string) (*pbAuthorizationV1.AuthorizationV1Policy, error)
	GetPolicies(ctx context.Context) ([]*pbAuthorizationV1.AuthorizationV1Policy, error)
	DeletePolicy(ctx context.Context, name string) error

	GetActions(ctx context.Context) ([]*pbAuthorizationV1.AuthorizationV1Action, error)
	GetAction(ctx context.Context, name string) (*pbAuthorizationV1.AuthorizationV1Action, error)

	EnsurePolicies(ctx context.Context, actions ...*pbAuthorizationV1.AuthorizationV1Policy) error
	EnsureActions(ctx context.Context, actions ...*pbAuthorizationV1.AuthorizationV1Action) error
	EnsureRoles(ctx context.Context, roles ...*pbAuthorizationV1.AuthorizationV1Role) error
}
