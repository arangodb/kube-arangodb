//
// DISCLAIMER
//
// Copyright 2025 ArangoDB GmbH, Cologne, Germany
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
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

func NewObject[IN proto.Message](in IN) Object[IN] {
	return Object[IN]{Object: in}
}

type Object[IN proto.Message] struct {
	Object IN
}

func (g *Object[T]) UnmarshalJSON(data []byte) error {
	return g.UnmarshalJSONOpts(data)
}

func (g *Object[T]) UnmarshalJSONOpts(data []byte, opts ...util.Mod[protojson.UnmarshalOptions]) error {
	o, err := Unmarshal[T](data, opts...)
	if err != nil {
		return err
	}

	g.Object = o
	return nil
}

func (g Object[T]) MarshalJSON() ([]byte, error) {
	return g.MarshalJSONOpts()
}

func (g Object[T]) MarshalJSONOpts(opts ...util.Mod[protojson.MarshalOptions]) ([]byte, error) {
	return Marshal[T](g.Object, opts...)
}
