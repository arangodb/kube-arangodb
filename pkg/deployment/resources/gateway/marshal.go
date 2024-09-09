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

package gateway

import (
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"sigs.k8s.io/yaml"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

func Marshal[T proto.Message](in T) ([]byte, string, T, error) {
	data, err := protojson.MarshalOptions{
		UseProtoNames: true,
	}.Marshal(in)
	if err != nil {
		return nil, "", util.Default[T](), err
	}

	data, err = yaml.JSONToYAML(data)
	return data, util.SHA256(data), in, err
}
