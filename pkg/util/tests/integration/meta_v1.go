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

package integration

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	pbMetaV1 "github.com/arangodb/kube-arangodb/integrations/meta/v1/definition"
	pbSharedV1 "github.com/arangodb/kube-arangodb/integrations/shared/v1/definition"
)

// NewInMemoryMetaV1Client returns an in-memory implementation of MetaV1Client for testing.
// Supports Get, Set (with revision-based precondition), Delete, and List with prefix filtering.
func NewMetaV1Client() pbMetaV1.MetaV1Client {
	return &inMemoryMetaStore{
		objects: make(map[string]*storedObject),
	}
}

type inMemoryMetaStore struct {
	lock    sync.Mutex
	objects map[string]*storedObject
	revSeq  int64
}

type storedObject struct {
	resp *pbMetaV1.ObjectResponse
}

func (m *inMemoryMetaStore) nextRev() string {
	m.revSeq++
	return fmt.Sprintf("_rev%d", m.revSeq)
}

func (m *inMemoryMetaStore) Get(ctx context.Context, in *pbMetaV1.ObjectRequest, opts ...grpc.CallOption) (*pbMetaV1.ObjectResponse, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	obj, ok := m.objects[in.Key]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "Key %s not found", in.Key)
	}

	return obj.resp, nil
}

func (m *inMemoryMetaStore) GetBatch(ctx context.Context, in *pbMetaV1.ObjectBatchRequest, opts ...grpc.CallOption) (*pbMetaV1.ObjectBatchResponse, error) {
	var items []*pbMetaV1.ObjectResponse
	for _, req := range in.Items {
		resp, err := m.Get(ctx, req, opts...)
		if err != nil {
			return nil, err
		}
		items = append(items, resp)
	}
	return &pbMetaV1.ObjectBatchResponse{Items: items}, nil
}

func (m *inMemoryMetaStore) Set(ctx context.Context, in *pbMetaV1.SetRequest, opts ...grpc.CallOption) (*pbMetaV1.ObjectResponse, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if rev := in.Revision; rev != nil {
		existing, ok := m.objects[in.Key]
		if !ok {
			return nil, status.Errorf(codes.NotFound, "Key %s not found", in.Key)
		}
		if existing.resp.GetRevision() != *rev {
			return nil, status.Errorf(codes.FailedPrecondition, "revision mismatch: expected %s, got %s", existing.resp.GetRevision(), *rev)
		}
	}

	newRev := m.nextRev()
	resp := &pbMetaV1.ObjectResponse{
		Key:      in.Key,
		Revision: &newRev,
		Object:   in.Object,
		Meta: &pbMetaV1.ObjectResponseMeta{
			Updated: timestamppb.Now(),
		},
	}

	m.objects[in.Key] = &storedObject{resp: resp}
	return resp, nil
}

func (m *inMemoryMetaStore) Delete(ctx context.Context, in *pbMetaV1.ObjectRequest, opts ...grpc.CallOption) (*pbSharedV1.Empty, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if _, ok := m.objects[in.Key]; !ok {
		return nil, status.Errorf(codes.NotFound, "Key %s not found", in.Key)
	}

	delete(m.objects, in.Key)
	return &pbSharedV1.Empty{}, nil
}

func (m *inMemoryMetaStore) List(ctx context.Context, in *pbMetaV1.ListRequest, opts ...grpc.CallOption) (pbMetaV1.MetaV1_ListClient, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	prefix := in.GetPrefix()
	var keys []string
	for k := range m.objects {
		if prefix == "" || strings.HasPrefix(k, prefix) {
			keys = append(keys, k)
		}
	}

	return &mockListClient{keys: keys}, nil
}

// mockListClient implements pbMetaV1.MetaV1_ListClient
type mockListClient struct {
	grpc.ClientStream
	keys []string
	sent bool
}

func (c *mockListClient) Recv() (*pbMetaV1.ListResponseChunk, error) {
	if c.sent {
		return nil, fmt.Errorf("EOF")
	}
	c.sent = true
	return &pbMetaV1.ListResponseChunk{Keys: c.keys}, nil
}
