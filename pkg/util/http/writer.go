//
// DISCLAIMER
//
// Copyright 2023-2025 ArangoDB GmbH, Cologne, Germany
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
	"io"
	goHttp "net/http"
)

func NewWriter(w goHttp.ResponseWriter, stream io.Writer) goHttp.ResponseWriter {
	return &writer{
		writer: w,
		stream: stream,
	}
}

type writer struct {
	writer goHttp.ResponseWriter
	stream io.Writer
}

func (w *writer) Write(bytes []byte) (int, error) {
	return w.stream.Write(bytes)
}

func (w *writer) WriteHeader(statusCode int) {
	w.writer.WriteHeader(statusCode)
}

func (w *writer) Header() goHttp.Header {
	return w.writer.Header()
}
