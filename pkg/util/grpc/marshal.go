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
	"os"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"sigs.k8s.io/yaml"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

func WithUseProtoNames(value bool) util.Mod[protojson.MarshalOptions] {
	return func(in *protojson.MarshalOptions) {
		in.UseProtoNames = value
	}
}

func WithEmitDefaultValues(value bool) util.Mod[protojson.MarshalOptions] {
	return func(in *protojson.MarshalOptions) {
		in.EmitDefaultValues = value
	}
}

func Marshal[T proto.Message](in T, opts ...util.Mod[protojson.MarshalOptions]) ([]byte, error) {
	options := protojson.MarshalOptions{}

	util.ApplyMods(&options, opts...)

	data, err := options.Marshal(in)
	if err != nil {
		return nil, err
	}

	return data, err
}

func MarshalYAML[T proto.Message](in T, opts ...util.Mod[protojson.MarshalOptions]) ([]byte, error) {
	data, err := Marshal[T](in, opts...)
	if err != nil {
		return nil, err
	}

	data, err = yaml.JSONToYAML(data)
	return data, err
}

func Unmarshal[T proto.Message](data []byte, opts ...util.Mod[protojson.UnmarshalOptions]) (T, error) {
	v, err := util.DeepType[T]()
	if err != nil {
		return util.Default[T](), err
	}

	options := protojson.UnmarshalOptions{}

	util.ApplyMods(&options, opts...)

	if err := options.Unmarshal(data, v); err != nil {
		return util.Default[T](), err
	}

	return v, nil
}

func UnmarshalFile[T proto.Message](path string, opts ...util.Mod[protojson.UnmarshalOptions]) (T, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return util.Default[T](), err
	}

	return Unmarshal[T](data, opts...)
}
