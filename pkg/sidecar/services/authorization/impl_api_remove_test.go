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
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pbImplAuthorizationV1 "github.com/arangodb/kube-arangodb/integrations/authorization/v1"
	sidecarSvcAuthzDefinition "github.com/arangodb/kube-arangodb/pkg/sidecar/services/authorization/definition"
	"github.com/arangodb/kube-arangodb/pkg/sidecar/services/authorization/pool"
	sidecarSvcAuthzTypes "github.com/arangodb/kube-arangodb/pkg/sidecar/services/authorization/types"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/svc/authenticator"
)

// stubPooler is a minimal pool.Pooler whose Delete returns a configurable error, used to
// assert the authorization API's error-code mapping on removal.
type stubPooler[T pool.PoolerObject] struct {
	deleteErr error
}

func (s *stubPooler[T]) Refresh(ctx context.Context) error { return nil }
func (s *stubPooler[T]) Create(ctx context.Context, name string, obj T) (T, uint32, error) {
	return obj, 0, nil
}
func (s *stubPooler[T]) Update(ctx context.Context, name string, obj T) (T, uint32, error) {
	return obj, 0, nil
}
func (s *stubPooler[T]) Delete(ctx context.Context, name string) (uint32, error) {
	return 0, s.deleteErr
}
func (s *stubPooler[T]) Item(name string) (T, uint32, bool)              { return util.Default[T](), 0, false }
func (s *stubPooler[T]) Index() uint32                                   { return 0 }
func (s *stubPooler[T]) Ready() bool                                     { return true }
func (s *stubPooler[T]) Pool(start uint32) ([]pool.OffsetItem[T], error) { return nil, nil }
func (s *stubPooler[T]) Offsets() []pool.OffsetItem[T]                   { return nil }
func (s *stubPooler[T]) Items() []string                                 { return nil }
func (s *stubPooler[T]) Copy() map[string]T                              { return nil }

func newRemoveTestImpl(bindingDeleteErr error) *implementation {
	return &implementation{
		policies:          &stubPooler[*sidecarSvcAuthzTypes.Policy]{},
		roles:             &stubPooler[*sidecarSvcAuthzTypes.Role]{},
		userRoleBindings:  &stubPooler[*sidecarSvcAuthzTypes.UserRoleBinding]{deleteErr: bindingDeleteErr},
		groupRoleBindings: &stubPooler[*sidecarSvcAuthzTypes.UserRoleBinding]{deleteErr: bindingDeleteErr},
		authType:          pbImplAuthorizationV1.ConfigurationTypeAlways,
	}
}

// Test_APIRemove_NotFound asserts that removing a role binding that does not exist returns
// NotFound (404) rather than Internal (500).
func Test_APIRemove_NotFound(t *testing.T) {
	ctx := authenticator.WithIdentity(context.Background(), &authenticator.Identity{User: util.NewType("admin")})

	t.Run("user role binding missing -> NotFound", func(t *testing.T) {
		a := newRemoveTestImpl(pool.PoolNotFound{})

		_, err := a.APIRemoveUserRole(ctx, &sidecarSvcAuthzDefinition.AuthorizationAPIUserRoleRequest{
			User: "alice",
			Role: "reader",
		})
		require.Equal(t, codes.NotFound, status.Code(err), "got: %v", err)
	})

	t.Run("group role binding missing -> NotFound", func(t *testing.T) {
		a := newRemoveTestImpl(pool.PoolNotFound{})

		_, err := a.APIRemoveGroupRole(ctx, &sidecarSvcAuthzDefinition.AuthorizationAPIGroupRoleRequest{
			Group: "admins",
			Role:  "reader",
		})
		require.Equal(t, codes.NotFound, status.Code(err), "got: %v", err)
	})

	t.Run("successful removal -> OK", func(t *testing.T) {
		a := newRemoveTestImpl(nil)

		_, err := a.APIRemoveUserRole(ctx, &sidecarSvcAuthzDefinition.AuthorizationAPIUserRoleRequest{
			User: "alice",
			Role: "reader",
		})
		require.NoError(t, err)
	})
}
