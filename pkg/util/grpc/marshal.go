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
	"sigs.k8s.io/yaml"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

func Marshal[T proto.Message](in T) ([]byte, error) {
	data, err := protojson.MarshalOptions{
		UseProtoNames: true,
	}.Marshal(in)
	if err != nil {
		return nil, err
	}

	return data, err
}

func MarshalYAML[T proto.Message](in T) ([]byte, error) {
	data, err := Marshal[T](in)
	if err != nil {
		return nil, err
	}

	data, err = yaml.JSONToYAML(data)
	return data, err
}

func Unmarshal[T proto.Message](data []byte) (T, error) {
	v, err := util.DeepType[T]()
	if err != nil {
		return util.Default[T](), err
	}

	if err := (protojson.UnmarshalOptions{}).Unmarshal(data, v); err != nil {
		return util.Default[T](), err
	}

	return v, nil
}
