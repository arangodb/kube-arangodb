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

package definition

import (
	"context"
	"fmt"
	"time"

	"github.com/arangodb/go-driver/v2/connection"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/cache"
	"github.com/arangodb/kube-arangodb/pkg/util/shutdown"
)

func NewRootRequestModifier(client AuthenticationV1Client) connection.Authentication {
	return authenticationModifier(NewRootFetcher(client))
}

type authenticationModifier cache.ObjectFetcher[string]

func (a authenticationModifier) RequestModifier(r connection.Request) error {
	ctx, c := context.WithTimeout(shutdown.Context(), time.Second)
	defer c()

	token, _, err := a(ctx)
	if err != nil {
		return err
	}

	if token != "" {
		r.AddHeader("Authorization", fmt.Sprintf("bearer %s", token))
	}

	return nil
}

func NewRootFetcher(client AuthenticationV1Client) cache.ObjectFetcher[string] {
	return NewFetcher(client, "root")
}

func NewFetcher(client AuthenticationV1Client, user string, roles ...string) cache.ObjectFetcher[string] {
	return func(ctx context.Context) (string, time.Duration, error) {
		resp, err := client.CreateToken(ctx, &CreateTokenRequest{
			User:  util.NewType(user),
			Roles: roles,
		})
		if err != nil {
			return "", 0, err
		}

		if lf := resp.GetLifetime(); lf != nil {
			return resp.GetToken(), lf.AsDuration() / 100 * 75, nil
		}

		return resp.GetToken(), 5 * time.Minute, nil
	}
}
