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

package gcs

import (
	"context"
	"fmt"
	"hash"
	"sync"

	"cloud.google.com/go/storage"
)

type writer struct {
	lock sync.Mutex

	done bool

	bytes    int64
	checksum hash.Hash

	write *storage.Writer
}

func (w *writer) Closed() bool {
	return w.done
}

func (w *writer) Write(p []byte) (int, error) {
	w.lock.Lock()
	defer w.lock.Unlock()

	n, err := w.write.Write(p)
	if err != nil {
		return 0, err
	}

	if n > 0 {
		w.bytes += int64(n)
		w.checksum.Write(p[:n])
	}

	return n, nil
}

func (w *writer) Close(ctx context.Context) (string, int64, error) {
	w.lock.Lock()
	defer w.lock.Unlock()

	if !w.done {
		if err := w.write.Close(); err != nil {
			return "", 0, err
		}
	}

	w.done = true

	return fmt.Sprintf("%02x", w.checksum.Sum(nil)), w.bytes, nil
}
