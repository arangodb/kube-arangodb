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
	"compress/gzip"
	goHttp "net/http"
)

const (
	EncodingAcceptHeader   = "Accept-Encoding"
	EncodingResponseHeader = "Content-Encoding"
)

func WithEncoding(in goHttp.HandlerFunc) goHttp.HandlerFunc {
	return func(writer goHttp.ResponseWriter, request *goHttp.Request) {
		encoding := request.Header.Values(EncodingAcceptHeader)
		request.Header.Del(EncodingAcceptHeader)

		method := ParseHeaders(encoding...).Accept("gzip", "identity")

		switch method {
		case "gzip":
			WithGZipEncoding(in)(writer, request)
		case "identity":
			WithIdentityEncoding(in)(writer, request)
		default:
			WithIdentityEncoding(in)(writer, request)
		}
	}
}

func WithIdentityEncoding(in goHttp.HandlerFunc) goHttp.HandlerFunc {
	return func(responseWriter goHttp.ResponseWriter, request *goHttp.Request) {
		in(responseWriter, request)
	}
}

func WithGZipEncoding(in goHttp.HandlerFunc) goHttp.HandlerFunc {
	return func(responseWriter goHttp.ResponseWriter, request *goHttp.Request) {
		responseWriter.Header().Add(EncodingResponseHeader, "gzip")

		stream := gzip.NewWriter(responseWriter)

		in(NewWriter(responseWriter, stream), request)

		if err := stream.Close(); err != nil {
			logger.Err(err).Warn("Unable to write GZIP response")
		}
	}
}
