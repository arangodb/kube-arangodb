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
	"io"
	"sync"
)

func wrapWithOffsetWriter(in io.Writer) io.WriterAt {
	return &offsetWriter{
		offset: 0,
		out:    in,
	}
}

type offsetWriter struct {
	lock   sync.Mutex
	offset int64
	out    io.Writer
}

func (o *offsetWriter) WriteAt(p []byte, off int64) (n int, err error) {
	o.lock.Lock()
	defer o.lock.Unlock()

	if o.offset != off {
		return 0, io.ErrUnexpectedEOF
	}

	n, err = o.out.Write(p)
	if err != nil {
		return 0, err
	}

	o.offset += int64(n)

	return n, nil
}
