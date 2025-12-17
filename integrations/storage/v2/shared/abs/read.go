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

package abs

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"hash"
	"io"
	"os"
	"sync"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"

	pbImplStorageV2Shared "github.com/arangodb/kube-arangodb/integrations/storage/v2/shared"
)

type reader struct {
	lock sync.Mutex

	in io.ReadCloser

	bytes    int64
	checksum hash.Hash

	closed bool
}

func (r *reader) Read(p []byte) (n int, err error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	n, err = r.in.Read(p)
	if n > 0 {
		r.bytes += int64(n)
		r.checksum.Write(p[:n])
	}

	if err != nil {
		if errors.Is(err, io.EOF) {
			r.closed = true
			if n > 0 {
				return n, nil
			}
			return
		}

		var respErr *azcore.ResponseError
		if errors.As(err, &respErr) && respErr.StatusCode == 404 {
			return 0, os.ErrNotExist
		}
		return 0, err
	}

	return
}

func (r *reader) Close(ctx context.Context) (string, int64, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	if err := r.in.Close(); err != nil {
		return "", 0, err
	}

	return fmt.Sprintf("%02x", r.checksum.Sum(nil)), r.bytes, nil
}

func (r *reader) Closed() bool {
	return r.closed
}

func (i *ios) Read(ctx context.Context, key string) (pbImplStorageV2Shared.Reader, error) {
	q := i.container().NewBlockBlobClient(i.key(key))

	resp, err := q.DownloadStream(ctx, nil)
	if err != nil {
		var respErr *azcore.ResponseError
		if errors.As(err, &respErr) && respErr.StatusCode == 404 {
			return nil, os.ErrNotExist
		}

		return nil, err
	}

	var reader reader

	reader.in = resp.Body
	reader.checksum = sha256.New()

	return &reader, nil
}
