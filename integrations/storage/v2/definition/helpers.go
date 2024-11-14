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
)

const BufferSize = 4094

func Send(ctx context.Context, client StorageV2Client, key string, in io.Reader) (*StorageV2WriteObjectResponse, error) {
	cache := make([]byte, BufferSize)

	wr, err := client.WriteObject(ctx)
	if err != nil {
		return nil, err
	}

	for {
		n, err := in.Read(cache)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			if cerr := wr.CloseSend(); cerr != nil {
				return nil, errors.Errors(err, cerr)
			}

			return nil, err
		}

		if err := wr.Send(&StorageV2WriteObjectRequest{
			Path: &StorageV2Path{
				Path: key,
			},
			Chunk: cache[:n],
		}); err != nil {
			if cerr := wr.CloseSend(); cerr != nil {
				return nil, errors.Errors(err, cerr)
			}

			return nil, err
		}
	}

	return wr.CloseAndRecv()
}

func Receive(ctx context.Context, client StorageV2Client, key string, out io.Writer) (int, error) {
	wr, err := client.ReadObject(ctx, &StorageV2ReadObjectRequest{
		Path: &StorageV2Path{Path: key},
	})
	if err != nil {
		return 0, err
	}

	var bytes int

	for {
		resp, err := wr.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			if cerr := wr.CloseSend(); cerr != nil {
				return 0, errors.Errors(err, cerr)
			}

			return 0, err
		}

		n, err := util.WriteAll(out, resp.GetChunk())
		if err != nil {
			if cerr := wr.CloseSend(); cerr != nil {
				return 0, errors.Errors(err, cerr)
			}

			return 0, err
		}

		bytes += n
	}

	return bytes, nil
}
