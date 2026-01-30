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
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/arangodb/kube-arangodb/integrations/meta/v1/definition"
	pbStorageV2 "github.com/arangodb/kube-arangodb/integrations/storage/v2/definition"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	tcache "github.com/arangodb/kube-arangodb/pkg/util/tests/cache"
)

func Test_Encryption(t *testing.T) {
	testEncryptionBatch(t, 1024)
	testEncryptionBatch(t, 1024*1024)
	testEncryptionBatch(t, 1024*1024*2)
}

func testEncryptionBatch(t *testing.T, size int) {
	t.Run(fmt.Sprintf("%d", size), func(t *testing.T) {
		ctx, c := context.WithCancel(context.Background())
		defer c()

		client := Client(t, tcache.NewRemoteCache[*Object](), ctx)

		obj := &pbStorageV2.StorageV2ReadObjectResponse{
			Chunk: make([]byte, size),
		}

		var ret pbStorageV2.StorageV2ReadObjectResponse

		any, err := anypb.New(obj)
		require.NoError(t, err)

		t.Run("Missing", func(t *testing.T) {
			resp, err := client.Get(ctx, &definition.ObjectRequest{Key: "/data"})
			require.EqualValues(t, codes.NotFound, errors.GRPCCode(err))
			require.Nil(t, resp)
		})

		t.Run("Set non-encrypted", func(t *testing.T) {
			_, err := client.Set(ctx, &definition.SetRequest{Key: "/data", Object: any})
			require.NoError(t, err)
		})

		t.Run("Get encrypted", func(t *testing.T) {
			resp, err := client.Get(ctx, &definition.ObjectRequest{Key: "/data", Secret: &definition.ObjectSecret{
				Secret: &definition.ObjectSecret_Token{Token: &definition.ObjectSecretToken{Token: "P@ssw0rd"}},
			}})
			require.EqualValues(t, codes.FailedPrecondition, errors.GRPCCode(err))
			require.EqualError(t, err, "rpc error: code = FailedPrecondition desc = Decryption failed: Object is not encrypted, but secret provided")
			require.Nil(t, resp)
		})

		t.Run("Get unencrypted", func(t *testing.T) {
			resp, err := client.Get(ctx, &definition.ObjectRequest{Key: "/data"})
			require.NoError(t, err)

			require.NoError(t, resp.Object.UnmarshalTo(&ret))

			require.Equal(t, util.SHA256(obj.Chunk), util.SHA256(ret.Chunk))
		})

		t.Run("Set encrypted", func(t *testing.T) {
			_, err = client.Set(ctx, &definition.SetRequest{Key: "/data", Object: any, Secret: &definition.ObjectSecret{
				Secret: &definition.ObjectSecret_Token{Token: &definition.ObjectSecretToken{Token: "P@ssw0rd"}},
			}})
			require.NoError(t, err)
		})

		_, err = client.Set(ctx, &definition.SetRequest{Key: "/data", Object: any, Secret: &definition.ObjectSecret{
			Secret: &definition.ObjectSecret_Token{Token: &definition.ObjectSecretToken{Token: "P@ssw0rd"}},
		}})
		require.NoError(t, err)

		t.Run("Get encrypted", func(t *testing.T) {
			resp, err := client.Get(ctx, &definition.ObjectRequest{Key: "/data", Secret: &definition.ObjectSecret{
				Secret: &definition.ObjectSecret_Token{Token: &definition.ObjectSecretToken{Token: "P@ssw0rd"}},
			}})
			require.NoError(t, err)
			require.NotNil(t, resp)

			require.NoError(t, resp.Object.UnmarshalTo(&ret))

			require.Equal(t, util.SHA256(obj.Chunk), util.SHA256(ret.Chunk))
		})

		t.Run("Get encrypted - invalid password", func(t *testing.T) {
			resp, err := client.Get(ctx, &definition.ObjectRequest{Key: "/data", Secret: &definition.ObjectSecret{
				Secret: &definition.ObjectSecret_Token{Token: &definition.ObjectSecretToken{Token: "P@ssw0rd2"}},
			}})
			require.EqualValues(t, codes.FailedPrecondition, errors.GRPCCode(err))
			require.EqualError(t, err, "rpc error: code = FailedPrecondition desc = Decryption failed: cipher: message authentication failed")
			require.Nil(t, resp)
		})

		t.Run("Get encrypted - missing password", func(t *testing.T) {
			resp, err := client.Get(ctx, &definition.ObjectRequest{Key: "/data"})
			require.EqualValues(t, codes.FailedPrecondition, errors.GRPCCode(err))
			require.EqualError(t, err, "rpc error: code = FailedPrecondition desc = Decryption failed: Object encrypted, but secret is missing")
			require.Nil(t, resp)
		})
	})
}
