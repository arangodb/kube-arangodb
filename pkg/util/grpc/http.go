//
// DISCLAIMER
//
// Copyright 2025-2026 ArangoDB GmbH, Cologne, Germany
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

package grpc

import (
	"context"
	"encoding/json"
	goHttp "net/http"

	rstatus "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"github.com/arangodb/kube-arangodb/pkg/util"
	operatorHTTP "github.com/arangodb/kube-arangodb/pkg/util/http"
)

func AsJSON[T proto.Message, O any](r operatorHTTP.Response[Object[T]]) (O, error) {
	data, err := r.Data()
	if err != nil {
		return util.Default[O](), err
	}

	var q O

	if err := json.Unmarshal(data, &q); err != nil {
		return util.Default[O](), err
	}

	return q, nil
}

func Get[T proto.Message](ctx context.Context, client operatorHTTP.HTTPClient, url string, mods ...util.Mod[goHttp.Request]) operatorHTTP.Response[Object[T]] {
	return operatorHTTP.Get[Object[T], *RequestError](ctx, client, url, mods...)
}

func Post[IN, T proto.Message](ctx context.Context, client operatorHTTP.HTTPClient, in IN, url string, mods ...util.Mod[goHttp.Request]) operatorHTTP.Response[Object[T]] {

	return operatorHTTP.Post[Object[IN], Object[T], *RequestError](ctx, client, NewObject(in), url, mods...)
}

type RequestError struct {
	err error
}

func (d *RequestError) UnmarshalJSON(i []byte) error {
	var st rstatus.Status

	if err := json.Unmarshal(i, &st); err != nil {
		d.err = status.Errorf(codes.Internal, "invalid grpc status json: %v", err)
		return nil
	}

	d.err = status.FromProto(&st).Err()

	return nil
}

func (d *RequestError) Error() string {
	return d.err.Error()
}
