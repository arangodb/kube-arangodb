//
// DISCLAIMER
//
// Copyright 2024-2025 ArangoDB GmbH, Cologne, Germany
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
	"encoding/json"
	"net/http"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

type ConfigDestinationStaticInterface interface {
	Validate() error
	StaticResponse() ([]byte, uint32, error)
}

type ConfigDestinationStaticMarshaller[T any] func(in T, opts ...util.Mod[protojson.MarshalOptions]) ([]byte, error)

type ConfigDestinationStatic[T any] struct {
	Code *uint32 `json:"insecure,omitempty"`

	Response T `json:"response,omitempty"`

	Marshaller ConfigDestinationStaticMarshaller[T] `json:"-"`

	Options []util.Mod[protojson.MarshalOptions]
}

func (c *ConfigDestinationStatic[T]) Validate() error {
	return nil
}

func (c *ConfigDestinationStatic[T]) StaticResponse() ([]byte, uint32, error) {
	data, err := c.Marshall()
	if err != nil {
		return nil, 0, err
	}

	return data, c.GetCode(), nil
}

func (c *ConfigDestinationStatic[T]) Marshall() ([]byte, error) {
	if c == nil || util.IsDefault(c.Response) {
		return []byte("{}"), nil
	}

	if m := c.Marshaller; m == nil {
		return json.Marshal(c.Response)
	} else {
		return m(c.Response)
	}
}

func (c *ConfigDestinationStatic[T]) GetCode() uint32 {
	if c == nil || c.Code == nil {
		return http.StatusOK
	}

	return *c.Code
}
