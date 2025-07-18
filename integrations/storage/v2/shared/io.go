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

package shared

import (
	"context"
	"io"
	"time"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

type Writer interface {
	io.Writer

	Close(ctx context.Context) (string, int64, error)

	Closed() bool
}

type Reader interface {
	io.Reader

	Close(ctx context.Context) (string, int64, error)

	Closed() bool
}

type File struct {
	Key  string
	Info Info
}

type Info struct {
	Size          uint64
	LastUpdatedAt time.Time
}

type IO interface {
	Init(ctx context.Context, opts *InitOptions) error
	Write(ctx context.Context, key string) (Writer, error)
	Read(ctx context.Context, key string) (Reader, error)
	Head(ctx context.Context, key string) (*Info, error)
	Delete(ctx context.Context, key string) (bool, error)
	List(ctx context.Context, key string) (util.NextIterator[[]File], error)
}

func ToIOReader(ctx context.Context, in Reader) io.ReadCloser {
	return ioReader{
		ctx:    ctx,
		reader: in,
	}
}

type ioReader struct {
	ctx context.Context

	reader Reader
}

func (i ioReader) Read(p []byte) (n int, err error) {
	return i.reader.Read(p)
}

func (i ioReader) Close() error {
	_, _, err := i.reader.Close(i.ctx)
	return err
}
