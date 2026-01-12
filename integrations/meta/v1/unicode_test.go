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

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/durationpb"

	pbMetaV1 "github.com/arangodb/kube-arangodb/integrations/meta/v1/definition"
	"github.com/arangodb/kube-arangodb/pkg/util"
	ugrpc "github.com/arangodb/kube-arangodb/pkg/util/grpc"
)

func GetAllKeys(t *testing.T, client pbMetaV1.MetaV1Client, req *pbMetaV1.ListRequest) []string {
	resp, err := client.List(t.Context(), req)
	require.NoError(t, err)

	all, err := ugrpc.RecvAll(resp)
	require.NoError(t, err)

	return util.FlattenList(util.FormatList(all, func(a *pbMetaV1.ListResponseChunk) []string {
		return a.GetKeys()
	}))
}

func Test_ArangoDBUnicode(t *testing.T) {
	ctx, c := context.WithCancel(context.Background())
	defer c()

	client := Client(t, GetInternalRemoteCache(t), ctx)

	keys := GetAllKeys(t, client, &pbMetaV1.ListRequest{})
	require.Len(t, keys, 0)

	_, err := client.Set(t.Context(), &pbMetaV1.SetRequest{
		Key:    "obj%32ect",
		Object: nil,
		Ttl:    durationpb.New(time.Second),
	})
	require.NoError(t, err)

	keys = GetAllKeys(t, client, &pbMetaV1.ListRequest{})
	require.Len(t, keys, 1)
	require.Contains(t, keys, "obj%32ect")

	_, err = client.Set(t.Context(), &pbMetaV1.SetRequest{
		Key:    "obj%32ect",
		Object: nil,
		Ttl:    durationpb.New(time.Second),
	})
	require.NoError(t, err)

	keys = GetAllKeys(t, client, &pbMetaV1.ListRequest{})
	require.Len(t, keys, 1)
	require.Contains(t, keys, "obj%32ect")

	_, err = client.Delete(t.Context(), &pbMetaV1.ObjectRequest{
		Key: "obj%32ect",
	})
	require.NoError(t, err)

	keys = GetAllKeys(t, client, &pbMetaV1.ListRequest{})
	require.Len(t, keys, 0)

	_, err = client.Set(t.Context(), &pbMetaV1.SetRequest{
		Key: "object.data.xyz",
	})
	require.NoError(t, err)

	keys = GetAllKeys(t, client, &pbMetaV1.ListRequest{})
	require.Len(t, keys, 1)
	require.Contains(t, keys, "object.data.xyz")

	_, err = client.Delete(t.Context(), &pbMetaV1.ObjectRequest{
		Key: "object.data.xyz",
	})
	require.NoError(t, err)

	keys = GetAllKeys(t, client, &pbMetaV1.ListRequest{})
	require.Len(t, keys, 0)
}
