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
	"fmt"
	"hash"
	"io"
	"sync"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blockblob"

	pbImplStorageV2Shared "github.com/arangodb/kube-arangodb/integrations/storage/v2/shared"
)

type writer struct {
	lock sync.Mutex

	in io.WriteCloser

	done chan struct{}

	bytes    int64
	checksum hash.Hash

	err error
}

func (w *writer) Write(p []byte) (n int, err error) {
	w.lock.Lock()
	defer w.lock.Unlock()

	n, err = w.in.Write(p)
	if err != nil {
		return 0, err
	}

	if n > 0 {
		w.bytes += int64(n)
		w.checksum.Write(p[:n])
	}

	return
}

func (w *writer) Close(ctx context.Context) (string, int64, error) {
	w.lock.Lock()
	defer w.lock.Unlock()

	if err := w.in.Close(); err != nil {
		return "", 0, err
	}

	<-w.done

	if w.err != nil {
		return "", 0, w.err
	}

	return fmt.Sprintf("%02x", w.checksum.Sum(nil)), w.bytes, nil
}

func (w *writer) Closed() bool {
	w.lock.Lock()
	defer w.lock.Unlock()

	select {
	case <-w.done:
		return true
	default:
		return false
	}
}

func (w *writer) run(ctx context.Context, client *blockblob.Client, data io.Reader) {
	defer close(w.done)

	_, err := client.UploadStream(ctx, data, &blockblob.UploadStreamOptions{
		BlockSize: 4 * 1024 * 1024,
	})
	w.err = err
}

func (i *ios) Write(ctx context.Context, key string) (pbImplStorageV2Shared.Writer, error) {
	q := i.container().NewBlockBlobClient(i.key(key))

	in, out := io.Pipe()

	var writer writer

	writer.done = make(chan struct{})
	writer.in = out
	writer.checksum = sha256.New()

	go writer.run(ctx, q, in)

	return &writer, nil
}
