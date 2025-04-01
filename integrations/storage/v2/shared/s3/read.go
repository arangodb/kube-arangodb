//
// DISCLAIMER
//
// Copyright 2024-2025 ArangoDB GmbH, Cologne, Germany
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

package s3

import (
	"context"
	"crypto/sha256"
	"fmt"
	"hash"
	"io"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func newReader(parent *ios) *reader {
	pr, pw := io.Pipe()

	return &reader{
		parent:   parent,
		closed:   make(chan string),
		pr:       pr,
		pw:       pw,
		checksum: sha256.New(),
	}
}

type reader struct {
	lock, closeLock sync.Mutex

	parent *ios

	closed chan string

	err error

	bytes    int64
	checksum hash.Hash

	pr io.Reader
	pw io.WriteCloser
}

func (w *reader) Read(p []byte) (n int, err error) {
	w.lock.Lock()
	defer w.lock.Unlock()

	n, err = w.pr.Read(p)
	if err == nil {
		w.bytes += int64(n)
		w.checksum.Write(p[:n])
		return n, nil
	}

	if errors.Is(err, io.EOF) {
		if !w.done() {
			return 0, io.ErrUnexpectedEOF
		}

		if IsAWSNotFoundError(w.err) {
			return 0, os.ErrNotExist
		}
	}

	return n, err
}

func (w *reader) Closed() bool {
	return w.done()
}

func (w *reader) Close(ctx context.Context) (string, int64, error) {
	w.lock.Lock()
	defer w.lock.Unlock()

	if !w.done() {
		return "", 0, io.ErrNoProgress
	}

	if err := w.err; err != nil {
		return "", 0, err
	}

	return fmt.Sprintf("%02x", w.checksum.Sum(nil)), w.bytes, nil
}

func (w *reader) done() bool {
	w.closeLock.Lock()
	defer w.closeLock.Unlock()

	select {
	case <-w.closed:
		return true
	default:
		return false
	}
}

func (w *reader) start(ctx context.Context, input *s3.GetObjectInput) {
	go w._start(ctx, input)
}

func (w *reader) _start(ctx context.Context, input *s3.GetObjectInput) {
	defer func() {
		w.closeLock.Lock()
		defer w.closeLock.Unlock()

		defer close(w.closed)

		if err := w.pw.Close(); err != nil {
			if w.err != nil {
				w.err = err
			}
		}

		buff := make([]byte, 128)

		for {
			_, err := w.pr.Read(buff)
			if err != nil {
				return
			}
		}
	}()

	_, err := w.parent.downloader.DownloadWithContext(ctx, wrapWithOffsetWriter(w.pw), input)
	if err != nil {
		w.err = err
		return
	}
}
