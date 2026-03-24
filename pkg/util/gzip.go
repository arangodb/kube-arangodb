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

package util

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"sync"
)

func NewGZipBuilder(out io.Writer) GZipBuilder {
	gz := gzip.NewWriter(out)

	tw := tar.NewWriter(gz)

	return &gzipBuilder{
		gzip: gz,
		tar:  tw,
	}
}

func GZipBuilderProcessTemplate[T any](t Template[T], obj T) GZipBuilderProcess {
	return func() ([]byte, error) {
		return t.RenderBytes(obj)
	}
}

func GZipBuilderProcessBytes(in []byte) GZipBuilderProcess {
	return func() ([]byte, error) {
		return in, nil
	}
}

type GZipBuilderProcess func() ([]byte, error)

type GZipBuilder interface {
	File(pc GZipBuilderProcess, path string, args ...any) GZipBuilder

	Done() error
}

type gzipBuilder struct {
	lock sync.Mutex

	gzip *gzip.Writer

	tar *tar.Writer
}

func (c *gzipBuilder) File(pc GZipBuilderProcess, path string, args ...any) GZipBuilder {
	c.lock.Lock()
	defer c.lock.Unlock()

	content, err := pc()
	if err != nil {
		return errorGZipBuilder{err: err}
	}

	if err := c.tar.WriteHeader(&tar.Header{
		Name: fmt.Sprintf(path, args...),
		Mode: 0644,
		Uid:  1000,
		Gid:  1000,
		Size: int64(len(content)),
	}); err != nil {
		return errorGZipBuilder{err: err}
	}

	if _, err := c.tar.Write(content); err != nil {
		return errorGZipBuilder{err: err}
	}

	return c
}

func (c *gzipBuilder) Done() error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if err := c.tar.Close(); err != nil {
		return err
	}

	if err := c.gzip.Close(); err != nil {
		return err
	}

	return nil
}

type errorGZipBuilder struct {
	err error
}

func (e errorGZipBuilder) File(pc GZipBuilderProcess, path string, args ...any) GZipBuilder {
	return e
}

func (e errorGZipBuilder) Done() error {
	return e.err
}
