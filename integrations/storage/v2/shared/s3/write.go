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

package s3

import (
	"context"
	"crypto/sha256"
	"fmt"
	"hash"
	"io"
	"sync"

	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func newWriter(parent *ios) *writer {
	pr, pw := io.Pipe()

	return &writer{
		parent:   parent,
		closed:   make(chan string),
		pr:       pr,
		pw:       pw,
		checksum: sha256.New(),
	}
}

type writer struct {
	lock sync.Mutex

	parent *ios

	closed chan string

	err error

	bytes    int64
	checksum hash.Hash

	pr io.Reader
	pw io.WriteCloser
}

func (w *writer) Closed() bool {
	return w.done()
}

func (w *writer) Write(p []byte) (n int, err error) {
	w.lock.Lock()
	defer w.lock.Unlock()

	if w.done() {
		return 0, w.err
	}

	n, err = w.pw.Write(p)
	if err != nil {
		return 0, err
	}

	if n > 0 {
		w.bytes += int64(n)
		w.checksum.Write(p[:n])
	}

	return n, nil
}

func (w *writer) done() bool {
	select {
	case <-w.closed:
		return true
	default:
		return false
	}
}

func (w *writer) Close(ctx context.Context) (string, int64, error) {
	w.lock.Lock()
	defer w.lock.Unlock()

	if !w.done() {
		if err := w.pw.Close(); err != nil {
			return "", 0, err
		}

		<-w.closed
	}

	if w.err != nil {
		return "", 0, w.err
	}

	return fmt.Sprintf("%02x", w.checksum.Sum(nil)), w.bytes, nil
}

func (w *writer) start(ctx context.Context, input *s3manager.UploadInput) {
	go w._start(ctx, input)
}

func (w *writer) _start(ctx context.Context, input *s3manager.UploadInput) {
	defer close(w.closed)

	defer func() {
		// Clean the channel

		buff := make([]byte, 128)

		for {
			_, err := w.pr.Read(buff)
			if err != nil {
				return
			}
		}
	}()

	input.Body = w.pr

	_, err := w.parent.uploader.UploadWithContext(ctx, input)
	if err != nil {
		w.err = err
		return
	}
}
