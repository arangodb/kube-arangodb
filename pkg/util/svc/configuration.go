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

package svc

import (
	"crypto/tls"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type Configuration struct {
	Address string

	TLSOptions *tls.Config

	Options []grpc.ServerOption

	Gateway *ConfigurationGateway
}

type ConfigurationGateway struct {
	Address string
}

func (c *Configuration) RenderOptions() []grpc.ServerOption {
	if c == nil {
		return nil
	}

	ret := make([]grpc.ServerOption, len(c.Options))
	copy(ret, c.Options)

	if tls := c.TLSOptions; tls != nil {
		ret = append(ret, grpc.Creds(credentials.NewTLS(tls)))
	}

	return ret
}
