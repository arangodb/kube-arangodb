//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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

package definition

import (
	"context"
	"io"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	ugrpc "github.com/arangodb/kube-arangodb/pkg/util/grpc"
)

const BufferSize = 4094

func Send(ctx context.Context, client StorageV2Client, key string, in io.Reader) (*StorageV2WriteObjectResponse, error) {
	cache := make([]byte, BufferSize)

	wr, err := client.WriteObject(ctx)
	if err != nil {
		return nil, err
	}

	return ugrpc.Send[*StorageV2WriteObjectRequest, *StorageV2WriteObjectResponse](wr, func() (*StorageV2WriteObjectRequest, error) {
		n, err := in.Read(cache)
		if err != nil {
			return nil, err
		}

		return &StorageV2WriteObjectRequest{
			Path: &StorageV2Path{
				Path: key,
			},
			Chunk: cache[:n],
		}, nil
	})
}

func Receive(ctx context.Context, client StorageV2Client, key string, out io.Writer) (int, error) {
	wr, err := client.ReadObject(ctx, &StorageV2ReadObjectRequest{
		Path: &StorageV2Path{Path: key},
	})
	if err != nil {
		return 0, err
	}

	var bytes int

	if err := ugrpc.Recv[*StorageV2ReadObjectResponse](wr, func(response *StorageV2ReadObjectResponse) error {
		n, err := util.WriteAll(out, response.GetChunk())
		if err != nil {
			if cerr := wr.CloseSend(); cerr != nil {
				return errors.Errors(err, cerr)
			}

			return err
		}

		bytes += n

		return nil
	}); err != nil {
		return 0, err
	}

	return bytes, nil
}
