//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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

package http

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

const (
	ContentLengthHeader = "Content-Length"
)

func WithBuffer(maxSize int, in http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		data := NewBuffer(maxSize, writer)

		wr := NewWriter(writer, data)

		in(wr, request)

		if !data.Truncated() {
			// We have constant size
			bytes := data.Bytes()

			println(len(bytes))

			writer.Header().Set(ContentLengthHeader, fmt.Sprintf("%d", len(bytes)))

			_, err := util.WriteAll(writer, bytes)
			if err != nil {
				logger.Err(err).Warn("Unable to write HTTP response")
			}
		}
	}
}

type Buffer interface {
	io.Writer

	Bytes() []byte
	Truncated() bool
}

type buffer struct {
	lock sync.Mutex

	upstream io.Writer

	data, currentData []byte
}

func (b *buffer) Write(q []byte) (n int, err error) {
	b.lock.Lock()
	defer b.lock.Unlock()

	p := q

	for {
		if len(p) == 0 || len(b.currentData) == 0 {
			break
		}

		b.currentData[0] = p[0]
		b.currentData = b.currentData[1:]
		p = p[1:]
	}

	if len(p) == 0 {
		return len(q), nil
	}

	written := 0

	if len(b.currentData) == 0 {
		if b.data != nil {
			z, err := util.WriteAll(b.upstream, b.data)
			if err != nil {
				return 0, err
			}

			written += z

			b.data = nil
			b.currentData = nil
		}
	} else {
		return len(q), nil
	}

	z, err := b.upstream.Write(p)
	if err != nil {
		return 0, err
	}

	written += z

	return written, nil
}

func (b *buffer) Bytes() []byte {
	b.lock.Lock()
	defer b.lock.Unlock()

	if len(b.data) == 0 {
		return nil
	}

	return b.data[:len(b.data)-len(b.currentData)]
}

func (b *buffer) Truncated() bool {
	b.lock.Lock()
	defer b.lock.Unlock()

	return len(b.data) == 0
}

type bytesBuffer struct {
	*bytes.Buffer
}

func (b bytesBuffer) Truncated() bool {
	return false
}

func NewBuffer(maxSize int, upstream io.Writer) Buffer {
	if maxSize <= 0 {
		return &bytesBuffer{
			Buffer: bytes.NewBuffer(nil),
		}
	}

	b := &buffer{
		data:     make([]byte, maxSize),
		upstream: upstream,
	}
	b.currentData = b.data
	return b
}
