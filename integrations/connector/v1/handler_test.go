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
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pbMetaV1 "github.com/arangodb/kube-arangodb/integrations/meta/v1/definition"
	testIntegration "github.com/arangodb/kube-arangodb/pkg/util/tests/integration"
)

func Test_Heartbeat_Registers(t *testing.T) {
	meta := testIntegration.NewMetaV1Client()
	connectorID := testConnectorID
	handlerID := uuid.New().String()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start heartbeat in background
	done := make(chan struct{})
	go func() {
		defer close(done)
		startHeartbeat(ctx, meta, connectorID, handlerID)
	}()

	// Give it time to register
	time.Sleep(100 * time.Millisecond)

	// Verify the handler key exists in MetaStore
	key := handlerKey(connectorID, handlerID)
	resp, err := meta.Get(context.Background(), &pbMetaV1.ObjectRequest{Key: key})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, key, resp.Key)

	cancel()
	<-done
}

func Test_Heartbeat_KeyFormat(t *testing.T) {
	connectorID := "11111111-1111-1111-1111-111111111111"
	handlerID := "22222222-2222-2222-2222-222222222222"

	key := handlerKey(connectorID, handlerID)
	require.Equal(t, "connectors/11111111-1111-1111-1111-111111111111/handlers/22222222-2222-2222-2222-222222222222", key)
}

func Test_Heartbeat_MultipleHandlers(t *testing.T) {
	meta := testIntegration.NewMetaV1Client()
	connectorID := testConnectorID
	handler1 := uuid.New().String()
	handler2 := uuid.New().String()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done1 := make(chan struct{})
	go func() {
		defer close(done1)
		startHeartbeat(ctx, meta, connectorID, handler1)
	}()

	done2 := make(chan struct{})
	go func() {
		defer close(done2)
		startHeartbeat(ctx, meta, connectorID, handler2)
	}()

	time.Sleep(100 * time.Millisecond)

	// Both handlers should be registered
	_, err := meta.Get(context.Background(), &pbMetaV1.ObjectRequest{Key: handlerKey(connectorID, handler1)})
	require.NoError(t, err)

	_, err = meta.Get(context.Background(), &pbMetaV1.ObjectRequest{Key: handlerKey(connectorID, handler2)})
	require.NoError(t, err)

	cancel()
	<-done1
	<-done2
}

func Test_Heartbeat_StopsOnCancel(t *testing.T) {
	meta := testIntegration.NewMetaV1Client()
	handlerID := uuid.New().String()

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		defer close(done)
		startHeartbeat(ctx, meta, testConnectorID, handlerID)
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// Good — heartbeat stopped
	case <-time.After(time.Second):
		t.Fatal("heartbeat did not stop after context cancellation")
	}
}

func Test_Heartbeat_NotRegisteredBeforeStart(t *testing.T) {
	meta := testIntegration.NewMetaV1Client()
	handlerID := uuid.New().String()

	key := handlerKey(testConnectorID, handlerID)
	_, err := meta.Get(context.Background(), &pbMetaV1.ObjectRequest{Key: key})
	require.Error(t, err)

	s, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.NotFound, s.Code())
}
