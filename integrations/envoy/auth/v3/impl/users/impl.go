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

package users

import (
	"context"
	"time"

	pbEnvoyAuthV3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	"k8s.io/apimachinery/pkg/util/uuid"

	adbDriverV2 "github.com/arangodb/go-driver/v2/arangodb"
	adbDriverV2Shared "github.com/arangodb/go-driver/v2/arangodb/shared"

	pbImplEnvoyAuthV3Shared "github.com/arangodb/kube-arangodb/integrations/envoy/auth/v3/shared"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/cache"
	utilConstantsContext "github.com/arangodb/kube-arangodb/pkg/util/constants/context"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func New(ctx context.Context, configuration pbImplEnvoyAuthV3Shared.Configuration) (pbImplEnvoyAuthV3Shared.AuthHandler, bool, error) {
	if !configuration.Enabled {
		return nil, false, nil
	}

	if !configuration.Extensions.UsersCreate {
		return nil, false, nil
	}

	i := &impl{}

	if c, ok := utilConstantsContext.ArangoDBClientCache.Get(ctx); ok {
		i.userClient = c
	} else {
		return nil, false, errors.Errorf("client not found")
	}

	i.users = cache.NewCache[string, adbDriverV2.User](func(ctx context.Context, in string) (adbDriverV2.User, time.Time, error) {
		client, err := i.userClient.Get(ctx)
		if err != nil {
			return nil, util.Default[time.Time](), err
		}

		if user, err := client.User(ctx, in); err == nil {
			return user, time.Now().Add(24 * time.Hour), nil
		} else {
			if !adbDriverV2Shared.IsNotFound(err) {
				return nil, util.Default[time.Time](), err
			}
		}

		if user, err := client.CreateUser(ctx, in, &adbDriverV2.UserOptions{
			Password: string(uuid.NewUUID()),
			Active:   util.NewType(true),
		}); err != nil {
			if !adbDriverV2Shared.IsConflict(err) {
				return nil, util.Default[time.Time](), err
			}
		} else {
			return user, time.Now().Add(24 * time.Hour), nil
		}

		v, err := client.User(ctx, in)
		return v, time.Now().Add(24 * time.Hour), err
	})

	return i, true, nil
}

type impl struct {
	userClient cache.Object[adbDriverV2.Client]

	users cache.Cache[string, adbDriverV2.User]
}

func (i *impl) Handle(ctx context.Context, request *pbEnvoyAuthV3.CheckRequest, current *pbImplEnvoyAuthV3Shared.Response) error {
	if !current.Authenticated() {
		return nil
	}

	_, err := i.users.Get(ctx, current.User.User)

	return err
}
