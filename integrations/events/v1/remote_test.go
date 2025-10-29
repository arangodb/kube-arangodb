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

package v1

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	ugrpc "github.com/arangodb/kube-arangodb/pkg/util/grpc"
)

type TestRemoteStore[IN proto.Message] interface {
	RemoteStore[IN]

	Events(t *testing.T) []IN
}

func NewArangoTestStore[IN proto.Message]() TestRemoteStore[IN] {
	return &testRemoteStore[IN]{}
}

type testRemoteStore[IN proto.Message] struct {
	lock sync.Mutex

	events [][]byte
}

func (r *testRemoteStore[IN]) Init(ctx context.Context) error {
	return nil
}

func (r *testRemoteStore[IN]) Events(t *testing.T) []IN {
	r.lock.Lock()
	defer r.lock.Unlock()

	var ret = make([]IN, len(r.events))

	for i, e := range r.events {
		v, err := ugrpc.Unmarshal[IN](e)
		require.NoError(t, err)
		ret[i] = v
	}

	logger.Int("size", len(ret)).Info("Fetched Events")

	return ret
}

func (r *testRemoteStore[IN]) Emit(ctx context.Context, events ...IN) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	var res = make([][]byte, len(events))

	logger.Info("Emitting events in testing")

	for id, z := range events {
		data, err := ugrpc.Marshal(z)
		if err != nil {
			return err
		}

		res[id] = data
	}

	logger.Int("size", len(res)).Info("Emitted events in testing")

	r.events = append(r.events, res...)
	return nil
}
