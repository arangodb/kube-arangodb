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

package nctx

import (
	"context"
	"io"
)

const RequestReadBytesKey Key = "operator.requestReadBytes"

func (c *Counter) WithRequestReadBytes(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	return context.WithValue(ctx, RequestReadBytesKey, c)
}

type requestReadBytes struct {
	c *Counter

	in io.Reader
}

func (r requestReadBytes) Read(p []byte) (n int, err error) {
	n, err = r.in.Read(p)
	r.c.add(uint64(n))
	return
}

func WithRequestReadBytes(ctx context.Context, reader io.Reader) io.Reader {
	v := ctx.Value(RequestReadBytesKey)
	if v == nil {
		return reader
	}

	z, ok := v.(*Counter)
	if !ok {
		return reader
	}

	return requestReadBytes{
		c:  z,
		in: reader,
	}
}
