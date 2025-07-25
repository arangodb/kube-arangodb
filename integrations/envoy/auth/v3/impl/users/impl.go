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

package users

import (
	"context"
	"time"

	pbEnvoyAuthV3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	"k8s.io/apimachinery/pkg/util/uuid"

	"github.com/arangodb/go-driver/v2/arangodb"
	"github.com/arangodb/go-driver/v2/arangodb/shared"

	pbImplEnvoyAuthV3Shared "github.com/arangodb/kube-arangodb/integrations/envoy/auth/v3/shared"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/cache"
)

func New(ctx context.Context, configuration pbImplEnvoyAuthV3Shared.Configuration) (pbImplEnvoyAuthV3Shared.AuthHandler, bool) {
	if !configuration.Enabled {
		return nil, false
	}

	if !configuration.Extensions.UsersCreate {
		return nil, false
	}

	i := &impl{}

	i.userClient = configuration.DatabaseClient(configuration.Endpoint)

	i.users = cache.NewCache[string, arangodb.User](func(ctx context.Context, in string) (arangodb.User, time.Time, error) {
		client, err := i.userClient.Get(ctx)
		if err != nil {
			return nil, util.Default[time.Time](), err
		}

		if user, err := client.User(ctx, in); err == nil {
			return user, time.Now().Add(24 * time.Hour), nil
		} else {
			if !shared.IsNotFound(err) {
				return nil, util.Default[time.Time](), err
			}
		}

		if user, err := client.CreateUser(ctx, in, &arangodb.UserOptions{
			Password: string(uuid.NewUUID()),
			Active:   util.NewType(true),
		}); err != nil {
			if !shared.IsConflict(err) {
				return nil, util.Default[time.Time](), err
			}
		} else {
			return user, time.Now().Add(24 * time.Hour), nil
		}

		v, err := client.User(ctx, in)
		return v, time.Now().Add(24 * time.Hour), err
	})

	return i, true
}

type impl struct {
	userClient cache.Object[arangodb.Client]

	users cache.Cache[string, arangodb.User]
}

func (i *impl) Handle(ctx context.Context, request *pbEnvoyAuthV3.CheckRequest, current *pbImplEnvoyAuthV3Shared.Response) error {
	if !current.Authenticated() {
		return nil
	}

	_, err := i.users.Get(ctx, current.User.User)

	return err
}
