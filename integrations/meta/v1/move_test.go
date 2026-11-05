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

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	pbMetaV1 "github.com/arangodb/kube-arangodb/integrations/meta/v1/definition"
	"github.com/arangodb/kube-arangodb/pkg/util/cache"
	tcache "github.com/arangodb/kube-arangodb/pkg/util/tests/cache"
)

func mustAny(t *testing.T, v string) *anypb.Any {
	a, err := anypb.New(wrapperspb.String(v))
	require.NoError(t, err)
	return a
}

// Test_Move exercises the gRPC Move handler and its outcome-to-status-code mapping
// against the in-memory cache, so it runs without a live ArangoDB.
func Test_Move(t *testing.T) {
	ctx, c := context.WithCancel(context.Background())
	defer c()

	client := Client(t, tcache.NewRemoteCache[*Object](), ctx)

	payload := mustAny(t, "payload")

	_, err := client.Set(ctx, &pbMetaV1.SetRequest{Key: "a", Object: payload})
	require.NoError(t, err)

	t.Run("Move preserves the payload under the new key", func(t *testing.T) {
		resp, err := client.Move(ctx, &pbMetaV1.MoveRequest{Source: "a", Destination: "b"})
		require.NoError(t, err)
		require.Equal(t, payload.GetTypeUrl(), resp.GetObject().GetTypeUrl())
		require.Equal(t, payload.GetValue(), resp.GetObject().GetValue())

		// Source is gone.
		_, err = client.Get(ctx, &pbMetaV1.ObjectRequest{Key: "a"})
		require.Equal(t, codes.NotFound, status.Code(err))

		// Destination holds the payload.
		got, err := client.Get(ctx, &pbMetaV1.ObjectRequest{Key: "b"})
		require.NoError(t, err)
		require.Equal(t, payload.GetValue(), got.GetObject().GetValue())
	})

	t.Run("Move of a missing source returns NotFound", func(t *testing.T) {
		_, err := client.Move(ctx, &pbMetaV1.MoveRequest{Source: "a", Destination: "x"})
		require.Equal(t, codes.NotFound, status.Code(err))
	})

	t.Run("Move onto an existing destination returns AlreadyExists", func(t *testing.T) {
		_, err := client.Set(ctx, &pbMetaV1.SetRequest{Key: "d", Object: mustAny(t, "other")})
		require.NoError(t, err)

		_, err = client.Move(ctx, &pbMetaV1.MoveRequest{Source: "b", Destination: "d"})
		require.Equal(t, codes.AlreadyExists, status.Code(err))

		// Both keys still exist.
		_, err = client.Get(ctx, &pbMetaV1.ObjectRequest{Key: "b"})
		require.NoError(t, err)
		_, err = client.Get(ctx, &pbMetaV1.ObjectRequest{Key: "d"})
		require.NoError(t, err)
	})
}

// Test_Move_Transaction exercises the transactional RemoteCache.Move against a live ArangoDB.
// It is skipped when TEST_ARANGODB_ENDPOINT is not set.
func Test_Move_Transaction(t *testing.T) {
	c := GetInternalRemoteCache(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	require.NoError(t, c.Init(ctx))

	put := func(key, payload string) {
		require.NoError(t, c.Put(ctx, key, &Object{
			Key:    key,
			Meta:   &ObjectMeta{Updated: meta.Now()},
			Object: ObjectProto{Object: mustAny(t, payload)},
		}))
	}

	t.Run("Move relocates the object atomically", func(t *testing.T) {
		put("src", "value")

		res, err := c.Move(ctx, "src", "dst", "")
		require.NoError(t, err)
		require.Equal(t, cache.MoveResultMoved, res)

		_, exists, err := c.Get(ctx, "src")
		require.NoError(t, err)
		require.False(t, exists)

		got, exists, err := c.Get(ctx, "dst")
		require.NoError(t, err)
		require.True(t, exists)
		require.Equal(t, []byte("value"), decodeString(t, got))
	})

	t.Run("Missing source", func(t *testing.T) {
		res, err := c.Move(ctx, "missing", "irrelevant", "")
		require.NoError(t, err)
		require.Equal(t, cache.MoveResultSourceNotFound, res)
	})

	t.Run("Revision precondition", func(t *testing.T) {
		put("rev-src", "value")

		obj, exists, err := c.Get(ctx, "rev-src")
		require.NoError(t, err)
		require.True(t, exists)
		require.NotEmpty(t, obj.GetRev())

		// Wrong revision -> conflict, nothing moved.
		res, err := c.Move(ctx, "rev-src", "rev-dst", "_wrong_rev")
		require.NoError(t, err)
		require.Equal(t, cache.MoveResultRevisionConflict, res)

		_, exists, err = c.Get(ctx, "rev-src")
		require.NoError(t, err)
		require.True(t, exists)

		_, exists, err = c.Get(ctx, "rev-dst")
		require.NoError(t, err)
		require.False(t, exists)

		// Correct revision -> moved.
		res, err = c.Move(ctx, "rev-src", "rev-dst", obj.GetRev())
		require.NoError(t, err)
		require.Equal(t, cache.MoveResultMoved, res)
	})

	t.Run("Destination exists", func(t *testing.T) {
		put("dup-src", "src-value")
		put("dup-dst", "dst-value")

		res, err := c.Move(ctx, "dup-src", "dup-dst", "")
		require.NoError(t, err)
		require.Equal(t, cache.MoveResultDestinationExists, res)

		// Both untouched.
		src, exists, err := c.Get(ctx, "dup-src")
		require.NoError(t, err)
		require.True(t, exists)
		require.Equal(t, []byte("src-value"), decodeString(t, src))

		dst, exists, err := c.Get(ctx, "dup-dst")
		require.NoError(t, err)
		require.True(t, exists)
		require.Equal(t, []byte("dst-value"), decodeString(t, dst))
	})
}

// decodeString extracts the wrapped StringValue payload from a stored Object.
func decodeString(t *testing.T, o *Object) []byte {
	var v wrapperspb.StringValue
	require.NoError(t, o.Object.Object.UnmarshalTo(&v))
	return []byte(v.GetValue())
}
