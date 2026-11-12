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
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pbAuthorizationV1 "github.com/arangodb/kube-arangodb/integrations/authorization/v1/definition"
	pbImplAuthorizationV1Shared "github.com/arangodb/kube-arangodb/integrations/authorization/v1/shared"
	pbMetaV1 "github.com/arangodb/kube-arangodb/integrations/meta/v1/definition"
	pbSharedV1 "github.com/arangodb/kube-arangodb/integrations/shared/v1/definition"
	sidecarSvcAuthzTypes "github.com/arangodb/kube-arangodb/pkg/sidecar/services/authorization/types"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/svc/authenticator"
)

// evalCall records one authorization evaluation (action + resource) the RBAC wrapper performed.
type evalCall struct {
	action   string
	resource string
}

// evalRecorder is a concurrency-safe log of evaluations (GetBatch evaluates items in parallel).
type evalRecorder struct {
	mu    sync.Mutex
	calls []evalCall
}

func (r *evalRecorder) add(action, resource string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls = append(r.calls, evalCall{action: action, resource: resource})
}

func (r *evalRecorder) snapshot() []evalCall {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]evalCall(nil), r.calls...)
}

// evaluator records every evaluation and allows it unless allow reports false (allow nil = allow all).
func evaluator(rec *evalRecorder, allow func(action, resource string) bool) pbImplAuthorizationV1Shared.Evaluator {
	return pbImplAuthorizationV1Shared.PluginFunc(func(_ context.Context, req *pbAuthorizationV1.AuthorizationV1PermissionRequest) (*pbAuthorizationV1.AuthorizationV1PermissionResponse, error) {
		if rec != nil {
			rec.add(req.GetAction(), req.GetResource())
		}
		if allow == nil || allow(req.GetAction(), req.GetResource()) {
			return &pbAuthorizationV1.AuthorizationV1PermissionResponse{Effect: sidecarSvcAuthzTypes.Effect_Allow}, nil
		}
		return &pbAuthorizationV1.AuthorizationV1PermissionResponse{Effect: sidecarSvcAuthzTypes.Effect_Deny, Message: "denied"}, nil
	})
}

func allowAll(rec *evalRecorder) pbImplAuthorizationV1Shared.Evaluator { return evaluator(rec, nil) }
func denyAll(rec *evalRecorder) pbImplAuthorizationV1Shared.Evaluator {
	return evaluator(rec, func(string, string) bool { return false })
}

// caller returns a context carrying a verified caller identity, as the server auth interceptor would.
func caller() context.Context {
	return authenticator.WithIdentity(context.Background(), &authenticator.Identity{User: util.NewType("alice")})
}

// fakeMeta is a minimal parent Interface: it records the operations the RBAC wrapper delegated to it and
// maps a key to a resource string via Key. Every other method is promoted from the nil embedded Interface
// and must not be called.
type fakeMeta struct {
	Interface

	mu         sync.Mutex
	getKeys    []string
	setKeys    []string
	deleteKeys []string
	moved      bool
	listed     bool
}

func (f *fakeMeta) Key(parts ...string) string { return "key:" + strings.Join(parts, "/") }

func (f *fakeMeta) Get(_ context.Context, req *pbMetaV1.ObjectRequest) (*pbMetaV1.ObjectResponse, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.getKeys = append(f.getKeys, req.GetKey())
	return &pbMetaV1.ObjectResponse{}, nil
}

func (f *fakeMeta) Set(_ context.Context, req *pbMetaV1.SetRequest) (*pbMetaV1.ObjectResponse, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.setKeys = append(f.setKeys, req.GetKey())
	return &pbMetaV1.ObjectResponse{}, nil
}

func (f *fakeMeta) Delete(_ context.Context, req *pbMetaV1.ObjectRequest) (*pbSharedV1.Empty, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.deleteKeys = append(f.deleteKeys, req.GetKey())
	return &pbSharedV1.Empty{}, nil
}

func (f *fakeMeta) Move(_ context.Context, _ *pbMetaV1.MoveRequest) (*pbMetaV1.ObjectResponse, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.moved = true
	return &pbMetaV1.ObjectResponse{}, nil
}

func (f *fakeMeta) List(_ *pbMetaV1.ListRequest, _ pbMetaV1.MetaV1_ListServer) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.listed = true
	return nil
}

func (f *fakeMeta) getCount() int    { f.mu.Lock(); defer f.mu.Unlock(); return len(f.getKeys) }
func (f *fakeMeta) setCount() int    { f.mu.Lock(); defer f.mu.Unlock(); return len(f.setKeys) }
func (f *fakeMeta) deleteCount() int { f.mu.Lock(); defer f.mu.Unlock(); return len(f.deleteKeys) }

// fakeListServer provides only the Context the RBAC wrapper reads to resolve the caller identity.
type fakeListServer struct {
	pbMetaV1.MetaV1_ListServer
	ctx context.Context
}

func (s fakeListServer) Context() context.Context { return s.ctx }

func newRBAC(parent Interface, authz pbImplAuthorizationV1Shared.Evaluator) *rbac {
	return &rbac{Interface: parent, authz: authz}
}

func Test_meta_rbac_Get(t *testing.T) {
	t.Run("Allow reaches the parent and evaluates meta:GetKey on the mapped key", func(t *testing.T) {
		parent := &fakeMeta{}
		rec := &evalRecorder{}
		r := newRBAC(parent, allowAll(rec))

		_, err := r.Get(caller(), &pbMetaV1.ObjectRequest{Key: "k"})
		require.NoError(t, err)
		require.Equal(t, []string{"k"}, parent.getKeys)
		require.Equal(t, []evalCall{{ActionGetKey, "key:k"}}, rec.snapshot())
	})

	t.Run("Deny blocks the parent", func(t *testing.T) {
		parent := &fakeMeta{}
		r := newRBAC(parent, denyAll(nil))

		_, err := r.Get(caller(), &pbMetaV1.ObjectRequest{Key: "k"})
		require.Equal(t, codes.PermissionDenied, status.Code(err))
		require.Equal(t, 0, parent.getCount())
	})

	t.Run("No verified caller is Unauthenticated and does not reach the parent", func(t *testing.T) {
		parent := &fakeMeta{}
		rec := &evalRecorder{}
		r := newRBAC(parent, allowAll(rec))

		_, err := r.Get(context.Background(), &pbMetaV1.ObjectRequest{Key: "k"})
		require.Equal(t, codes.Unauthenticated, status.Code(err))
		require.Equal(t, 0, parent.getCount())
		require.Empty(t, rec.snapshot(), "evaluator must not be consulted without a caller")
	})
}

func Test_meta_rbac_Set(t *testing.T) {
	t.Run("Allow evaluates meta:UpdateKey and reaches the parent", func(t *testing.T) {
		parent := &fakeMeta{}
		rec := &evalRecorder{}
		r := newRBAC(parent, allowAll(rec))

		_, err := r.Set(caller(), &pbMetaV1.SetRequest{Key: "k"})
		require.NoError(t, err)
		require.Equal(t, []string{"k"}, parent.setKeys)
		require.Equal(t, []evalCall{{ActionUpdateKey, "key:k"}}, rec.snapshot())
	})

	t.Run("Deny blocks the parent", func(t *testing.T) {
		parent := &fakeMeta{}
		r := newRBAC(parent, denyAll(nil))

		_, err := r.Set(caller(), &pbMetaV1.SetRequest{Key: "k"})
		require.Equal(t, codes.PermissionDenied, status.Code(err))
		require.Equal(t, 0, parent.setCount())
	})
}

func Test_meta_rbac_Delete(t *testing.T) {
	t.Run("Allow evaluates meta:DeleteKey and reaches the parent", func(t *testing.T) {
		parent := &fakeMeta{}
		rec := &evalRecorder{}
		r := newRBAC(parent, allowAll(rec))

		_, err := r.Delete(caller(), &pbMetaV1.ObjectRequest{Key: "k"})
		require.NoError(t, err)
		require.Equal(t, []string{"k"}, parent.deleteKeys)
		require.Equal(t, []evalCall{{ActionDeleteKey, "key:k"}}, rec.snapshot())
	})

	t.Run("Deny blocks the parent", func(t *testing.T) {
		parent := &fakeMeta{}
		r := newRBAC(parent, denyAll(nil))

		_, err := r.Delete(caller(), &pbMetaV1.ObjectRequest{Key: "k"})
		require.Equal(t, codes.PermissionDenied, status.Code(err))
		require.Equal(t, 0, parent.deleteCount())
	})
}

func Test_meta_rbac_Move(t *testing.T) {
	t.Run("Allow evaluates meta:UpdateKey on both source and destination", func(t *testing.T) {
		parent := &fakeMeta{}
		rec := &evalRecorder{}
		r := newRBAC(parent, allowAll(rec))

		_, err := r.Move(caller(), &pbMetaV1.MoveRequest{Source: "s", Destination: "d"})
		require.NoError(t, err)
		require.True(t, parent.moved)
		require.Equal(t, []evalCall{{ActionUpdateKey, "key:s"}, {ActionUpdateKey, "key:d"}}, rec.snapshot())
	})

	t.Run("Deny on the destination key blocks the move", func(t *testing.T) {
		parent := &fakeMeta{}
		r := newRBAC(parent, evaluator(nil, func(_, resource string) bool { return resource != "key:d" }))

		_, err := r.Move(caller(), &pbMetaV1.MoveRequest{Source: "s", Destination: "d"})
		require.Equal(t, codes.PermissionDenied, status.Code(err))
		require.False(t, parent.moved)
	})
}

func Test_meta_rbac_List(t *testing.T) {
	t.Run("Allow evaluates meta:ListKey on the raw prefix and reaches the parent", func(t *testing.T) {
		parent := &fakeMeta{}
		rec := &evalRecorder{}
		r := newRBAC(parent, allowAll(rec))

		err := r.List(&pbMetaV1.ListRequest{Prefix: util.NewType("pre/")}, fakeListServer{ctx: caller()})
		require.NoError(t, err)
		require.True(t, parent.listed)
		require.Equal(t, []evalCall{{ActionListKey, "pre/"}}, rec.snapshot())
	})

	t.Run("Deny blocks the parent", func(t *testing.T) {
		parent := &fakeMeta{}
		r := newRBAC(parent, denyAll(nil))

		err := r.List(&pbMetaV1.ListRequest{Prefix: util.NewType("pre/")}, fakeListServer{ctx: caller()})
		require.Equal(t, codes.PermissionDenied, status.Code(err))
		require.False(t, parent.listed)
	})
}

func Test_meta_rbac_GetBatch(t *testing.T) {
	t.Run("Allow evaluates meta:GetKey per item and reaches the parent for each", func(t *testing.T) {
		parent := &fakeMeta{}
		rec := &evalRecorder{}
		r := newRBAC(parent, allowAll(rec))

		_, err := r.GetBatch(caller(), &pbMetaV1.ObjectBatchRequest{Items: []*pbMetaV1.ObjectRequest{
			{Key: "a"}, {Key: "b"}, {Key: "c"},
		}})
		require.NoError(t, err)
		require.ElementsMatch(t, []string{"a", "b", "c"}, parent.getKeys)
		require.ElementsMatch(t, []evalCall{{ActionGetKey, "key:a"}, {ActionGetKey, "key:b"}, {ActionGetKey, "key:c"}}, rec.snapshot())
	})

	t.Run("A single denied item fails the batch", func(t *testing.T) {
		parent := &fakeMeta{}
		r := newRBAC(parent, evaluator(nil, func(_, resource string) bool { return resource != "key:b" }))

		_, err := r.GetBatch(caller(), &pbMetaV1.ObjectBatchRequest{Items: []*pbMetaV1.ObjectRequest{
			{Key: "a"}, {Key: "b"},
		}})
		// GetBatch fans the per-item Get out through ParallelProcessOutputErr, which wraps the first error, so
		// the gRPC status code is not preserved on the batch error; the denied item still fails the batch.
		require.ErrorContains(t, err, "denied")
	})
}
