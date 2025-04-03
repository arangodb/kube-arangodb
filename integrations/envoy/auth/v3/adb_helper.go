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

package v3

import (
	"context"
	"sync"
	"time"

	"google.golang.org/protobuf/types/known/durationpb"

	pbAuthenticationV1 "github.com/arangodb/kube-arangodb/integrations/authentication/v1/definition"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/cache"
)

const (
	DefaultLifetime = time.Minute * 5
	DefaultTTL      = time.Minute
)

func NewADBHelper(client pbAuthenticationV1.AuthenticationV1Client) ADBHelper {
	return &adbHelper{
		client: client,
		cache: cache.NewCache(func(ctx context.Context, in string) (*pbAuthenticationV1.CreateTokenResponse, error) {
			return client.CreateToken(ctx, &pbAuthenticationV1.CreateTokenRequest{
				Lifetime: durationpb.New(DefaultLifetime),
				User:     util.NewType(in),
			})
		}, DefaultTTL),
	}
}

type ADBHelper interface {
	Validate(ctx context.Context, token string) (*AuthResponse, error)
	Token(ctx context.Context, resp *AuthResponse) (string, bool, error)
}

type adbHelper struct {
	lock  sync.Mutex
	cache cache.Cache[string, *pbAuthenticationV1.CreateTokenResponse]

	client pbAuthenticationV1.AuthenticationV1Client
}

func (a *adbHelper) Token(ctx context.Context, resp *AuthResponse) (string, bool, error) {
	a.lock.Lock()
	defer a.lock.Unlock()

	if resp == nil {
		// Token cannot be fetch if authentication is not valid
		return "", false, nil
	}

	token, err := a.cache.Get(ctx, resp.Username)
	if err != nil {
		return "", false, err
	}

	return token.GetToken(), true, nil
}

func (a *adbHelper) Validate(ctx context.Context, token string) (*AuthResponse, error) {
	a.lock.Lock()
	defer a.lock.Unlock()

	resp, err := a.client.Validate(ctx, &pbAuthenticationV1.ValidateRequest{
		Token: token,
	})

	if err != nil {
		return nil, err
	}

	if resp.GetIsValid() {
		if det := resp.GetDetails(); det != nil {
			return &AuthResponse{
				Username: det.GetUser(),
				Token:    util.NewType(token),
			}, nil
		}
	}

	return nil, nil
}
