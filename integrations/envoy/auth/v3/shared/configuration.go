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

package shared

import (
	"context"
	"time"

	pbAuthenticationV1 "github.com/arangodb/kube-arangodb/integrations/authentication/v1/definition"
	ugrpc "github.com/arangodb/kube-arangodb/pkg/util/grpc"
)

type Configuration struct {
	Address string

	Database ConfigurationDatabase

	Extensions ConfigurationExtensions

	Auth ConfigurationAuth
}

type ConfigurationDatabase struct {
	Proto    string
	Endpoint string
	Port     int
}

type ConfigurationAuth struct {
	Enabled bool
	Type    string
	Path    string
}

type ConfigurationExtensions struct {
	JWT         bool
	CookieJWT   bool
	UsersCreate bool
}

func (c Configuration) GetAuthClientFetcher(ctx context.Context) (pbAuthenticationV1.AuthenticationV1Client, time.Duration, error) {
	client, _, err := ugrpc.NewGRPCClient(ctx, pbAuthenticationV1.NewAuthenticationV1Client, c.Address)
	if err != nil {
		return nil, 0, err
	}

	return client, time.Hour, nil
}
