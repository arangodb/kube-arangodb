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

	"github.com/arangodb/kube-arangodb/integrations/authorization/v1/definition"
)

type Storage interface {
	Init(ctx context.Context) error
	Clean(ctx context.Context) error

	GetRole(ctx context.Context, name string) (*definition.AuthorizationV1Role, error)
	GetRoles(ctx context.Context) ([]*definition.AuthorizationV1Role, error)
	DeleteRole(ctx context.Context, name string) error

	RolePolicies(ctx context.Context, name string) ([]*definition.AuthorizationV1Policy, error)
	AssignRoleToPolicy(ctx context.Context, role, policy string) error

	GetPolicy(ctx context.Context, name string) (*definition.AuthorizationV1Policy, error)
	GetPolicies(ctx context.Context) ([]*definition.AuthorizationV1Policy, error)
	DeletePolicy(ctx context.Context, name string) error

	GetActions(ctx context.Context) ([]*definition.AuthorizationV1Action, error)
	GetAction(ctx context.Context, name string) (*definition.AuthorizationV1Action, error)

	EnsurePolicies(ctx context.Context, actions ...*definition.AuthorizationV1Policy) error
	EnsureActions(ctx context.Context, actions ...*definition.AuthorizationV1Action) error
	EnsureRoles(ctx context.Context, roles ...*definition.AuthorizationV1Role) error
}
