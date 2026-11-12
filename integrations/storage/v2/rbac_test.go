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

package v2

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pbAuthorizationV1 "github.com/arangodb/kube-arangodb/integrations/authorization/v1/definition"
	pbImplAuthorizationV1Shared "github.com/arangodb/kube-arangodb/integrations/authorization/v1/shared"
	pbStorageV2 "github.com/arangodb/kube-arangodb/integrations/storage/v2/definition"
	sidecarSvcAuthzTypes "github.com/arangodb/kube-arangodb/pkg/sidecar/services/authorization/types"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/svc/authenticator"
)

// evalCall records one authorization evaluation (action + resource) the RBAC wrapper performed.
type evalCall struct {
	action   string
	resource string
}

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

func path(p string) *pbStorageV2.StorageV2Path { return &pbStorageV2.StorageV2Path{Path: p} }

func newRBAC(parent Interface, authz pbImplAuthorizationV1Shared.Evaluator) *rbac {
	return &rbac{Interface: parent, authz: authz}
}

// fakeStorage is a minimal parent Interface recording the operations the RBAC wrapper delegated to it.
// Every other method is promoted from the nil embedded Interface and must not be called.
type fakeStorage struct {
	Interface

	mu     sync.Mutex
	head   []string
	del    []string
	read   []string
	list   []string
	inited bool
	wrote  bool
}

func (f *fakeStorage) HeadObject(_ context.Context, req *pbStorageV2.StorageV2HeadObjectRequest) (*pbStorageV2.StorageV2HeadObjectResponse, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.head = append(f.head, req.GetPath().GetPath())
	return &pbStorageV2.StorageV2HeadObjectResponse{}, nil
}

func (f *fakeStorage) DeleteObject(_ context.Context, req *pbStorageV2.StorageV2DeleteObjectRequest) (*pbStorageV2.StorageV2DeleteObjectResponse, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.del = append(f.del, req.GetPath().GetPath())
	return &pbStorageV2.StorageV2DeleteObjectResponse{}, nil
}

func (f *fakeStorage) Init(_ context.Context, _ *pbStorageV2.StorageV2InitRequest) (*pbStorageV2.StorageV2InitResponse, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.inited = true
	return &pbStorageV2.StorageV2InitResponse{}, nil
}

func (f *fakeStorage) ReadObject(req *pbStorageV2.StorageV2ReadObjectRequest, _ pbStorageV2.StorageV2_ReadObjectServer) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.read = append(f.read, req.GetPath().GetPath())
	return nil
}

func (f *fakeStorage) ListObjects(req *pbStorageV2.StorageV2ListObjectsRequest, _ pbStorageV2.StorageV2_ListObjectsServer) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.list = append(f.list, req.GetPath().GetPath())
	return nil
}

func (f *fakeStorage) WriteObject(_ pbStorageV2.StorageV2_WriteObjectServer) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.wrote = true
	return nil
}

func (f *fakeStorage) headCount() int { f.mu.Lock(); defer f.mu.Unlock(); return len(f.head) }
func (f *fakeStorage) delCount() int  { f.mu.Lock(); defer f.mu.Unlock(); return len(f.del) }
func (f *fakeStorage) readCount() int { f.mu.Lock(); defer f.mu.Unlock(); return len(f.read) }
func (f *fakeStorage) listCount() int { f.mu.Lock(); defer f.mu.Unlock(); return len(f.list) }

// fake stream servers provide only the Context the RBAC wrapper reads to resolve the caller identity (and,
// for writes, the first peeked message).
type fakeReadServer struct {
	pbStorageV2.StorageV2_ReadObjectServer
	ctx context.Context
}

func (s fakeReadServer) Context() context.Context { return s.ctx }

type fakeListServer struct {
	pbStorageV2.StorageV2_ListObjectsServer
	ctx context.Context
}

func (s fakeListServer) Context() context.Context { return s.ctx }

type fakeWriteServer struct {
	pbStorageV2.StorageV2_WriteObjectServer
	ctx      context.Context
	first    *pbStorageV2.StorageV2WriteObjectRequest
	firstErr error
}

func (s *fakeWriteServer) Context() context.Context { return s.ctx }
func (s *fakeWriteServer) Recv() (*pbStorageV2.StorageV2WriteObjectRequest, error) {
	return s.first, s.firstErr
}

func Test_storage_rbac_HeadObject(t *testing.T) {
	t.Run("Allow evaluates storage:HeadObject on the path and reaches the parent", func(t *testing.T) {
		parent := &fakeStorage{}
		rec := &evalRecorder{}
		r := newRBAC(parent, allowAll(rec))

		_, err := r.HeadObject(caller(), &pbStorageV2.StorageV2HeadObjectRequest{Path: path("a/b")})
		require.NoError(t, err)
		require.Equal(t, []string{"a/b"}, parent.head)
		require.Equal(t, []evalCall{{ActionHeadObject, "a/b"}}, rec.snapshot())
	})

	t.Run("Deny blocks the parent", func(t *testing.T) {
		parent := &fakeStorage{}
		r := newRBAC(parent, denyAll(nil))

		_, err := r.HeadObject(caller(), &pbStorageV2.StorageV2HeadObjectRequest{Path: path("a/b")})
		require.Equal(t, codes.PermissionDenied, status.Code(err))
		require.Equal(t, 0, parent.headCount())
	})

	t.Run("No verified caller is Unauthenticated", func(t *testing.T) {
		parent := &fakeStorage{}
		r := newRBAC(parent, allowAll(nil))

		_, err := r.HeadObject(context.Background(), &pbStorageV2.StorageV2HeadObjectRequest{Path: path("a/b")})
		require.Equal(t, codes.Unauthenticated, status.Code(err))
		require.Equal(t, 0, parent.headCount())
	})
}

func Test_storage_rbac_DeleteObject(t *testing.T) {
	t.Run("Allow evaluates storage:DeleteObject and reaches the parent", func(t *testing.T) {
		parent := &fakeStorage{}
		rec := &evalRecorder{}
		r := newRBAC(parent, allowAll(rec))

		_, err := r.DeleteObject(caller(), &pbStorageV2.StorageV2DeleteObjectRequest{Path: path("a/b")})
		require.NoError(t, err)
		require.Equal(t, []string{"a/b"}, parent.del)
		require.Equal(t, []evalCall{{ActionDeleteObject, "a/b"}}, rec.snapshot())
	})

	t.Run("Deny blocks the parent", func(t *testing.T) {
		parent := &fakeStorage{}
		r := newRBAC(parent, denyAll(nil))

		_, err := r.DeleteObject(caller(), &pbStorageV2.StorageV2DeleteObjectRequest{Path: path("a/b")})
		require.Equal(t, codes.PermissionDenied, status.Code(err))
		require.Equal(t, 0, parent.delCount())
	})
}

func Test_storage_rbac_Init(t *testing.T) {
	t.Run("Allow evaluates storage:Init with an empty resource and reaches the parent", func(t *testing.T) {
		parent := &fakeStorage{}
		rec := &evalRecorder{}
		r := newRBAC(parent, allowAll(rec))

		_, err := r.Init(caller(), &pbStorageV2.StorageV2InitRequest{})
		require.NoError(t, err)
		require.True(t, parent.inited)
		require.Equal(t, []evalCall{{ActionInit, ""}}, rec.snapshot())
	})

	t.Run("Deny blocks the parent", func(t *testing.T) {
		parent := &fakeStorage{}
		r := newRBAC(parent, denyAll(nil))

		_, err := r.Init(caller(), &pbStorageV2.StorageV2InitRequest{})
		require.Equal(t, codes.PermissionDenied, status.Code(err))
		require.False(t, parent.inited)
	})
}

func Test_storage_rbac_ReadObject(t *testing.T) {
	t.Run("Allow evaluates storage:ReadObject and reaches the parent", func(t *testing.T) {
		parent := &fakeStorage{}
		rec := &evalRecorder{}
		r := newRBAC(parent, allowAll(rec))

		err := r.ReadObject(&pbStorageV2.StorageV2ReadObjectRequest{Path: path("a/b")}, fakeReadServer{ctx: caller()})
		require.NoError(t, err)
		require.Equal(t, []string{"a/b"}, parent.read)
		require.Equal(t, []evalCall{{ActionReadObject, "a/b"}}, rec.snapshot())
	})

	t.Run("Deny blocks the parent", func(t *testing.T) {
		parent := &fakeStorage{}
		r := newRBAC(parent, denyAll(nil))

		err := r.ReadObject(&pbStorageV2.StorageV2ReadObjectRequest{Path: path("a/b")}, fakeReadServer{ctx: caller()})
		require.Equal(t, codes.PermissionDenied, status.Code(err))
		require.Equal(t, 0, parent.readCount())
	})
}

func Test_storage_rbac_ListObjects(t *testing.T) {
	t.Run("Allow evaluates storage:ListObjects and reaches the parent", func(t *testing.T) {
		parent := &fakeStorage{}
		rec := &evalRecorder{}
		r := newRBAC(parent, allowAll(rec))

		err := r.ListObjects(&pbStorageV2.StorageV2ListObjectsRequest{Path: path("pre/")}, fakeListServer{ctx: caller()})
		require.NoError(t, err)
		require.Equal(t, []string{"pre/"}, parent.list)
		require.Equal(t, []evalCall{{ActionListObjects, "pre/"}}, rec.snapshot())
	})

	t.Run("Deny blocks the parent", func(t *testing.T) {
		parent := &fakeStorage{}
		r := newRBAC(parent, denyAll(nil))

		err := r.ListObjects(&pbStorageV2.StorageV2ListObjectsRequest{Path: path("pre/")}, fakeListServer{ctx: caller()})
		require.Equal(t, codes.PermissionDenied, status.Code(err))
		require.Equal(t, 0, parent.listCount())
	})
}

func Test_storage_rbac_WriteObject(t *testing.T) {
	t.Run("Allow evaluates storage:WriteObject on the first message path and reaches the parent", func(t *testing.T) {
		parent := &fakeStorage{}
		rec := &evalRecorder{}
		r := newRBAC(parent, allowAll(rec))

		err := r.WriteObject(&fakeWriteServer{ctx: caller(), first: &pbStorageV2.StorageV2WriteObjectRequest{Path: path("w/obj")}})
		require.NoError(t, err)
		require.True(t, parent.wrote)
		require.Equal(t, []evalCall{{ActionWriteObject, "w/obj"}}, rec.snapshot())
	})

	t.Run("Deny blocks the parent", func(t *testing.T) {
		parent := &fakeStorage{}
		r := newRBAC(parent, denyAll(nil))

		err := r.WriteObject(&fakeWriteServer{ctx: caller(), first: &pbStorageV2.StorageV2WriteObjectRequest{Path: path("w/obj")}})
		require.Equal(t, codes.PermissionDenied, status.Code(err))
		require.False(t, parent.wrote)
	})

	t.Run("A first message with no path is not evaluated and reaches the parent", func(t *testing.T) {
		parent := &fakeStorage{}
		rec := &evalRecorder{}
		r := newRBAC(parent, allowAll(rec))

		err := r.WriteObject(&fakeWriteServer{ctx: caller(), first: &pbStorageV2.StorageV2WriteObjectRequest{}})
		require.NoError(t, err)
		require.True(t, parent.wrote)
		require.Empty(t, rec.snapshot(), "an empty path must not trigger an authorization check")
	})
}
