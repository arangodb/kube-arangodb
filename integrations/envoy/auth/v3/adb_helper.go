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

package v3

import (
	"context"
	"sync"
	"time"

	"google.golang.org/protobuf/types/known/durationpb"

	pbAuthenticationV1 "github.com/arangodb/kube-arangodb/integrations/authentication/v1/definition"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

const (
	DefaultLifetime = time.Minute * 5
	DefaultTTL      = time.Minute
)

func NewADBHelper(client pbAuthenticationV1.AuthenticationV1Client) ADBHelper {
	return &adbHelper{
		client: client,
		cache:  map[string]adbHelperToken{},
	}
}

type ADBHelper interface {
	Validate(ctx context.Context, token string) (*AuthResponse, error)
	Token(ctx context.Context, resp *AuthResponse) (string, bool, error)
}

type adbHelperToken struct {
	TTL   time.Time
	Token string
}

type adbHelper struct {
	lock  sync.Mutex
	cache map[string]adbHelperToken

	client pbAuthenticationV1.AuthenticationV1Client
}

func (a *adbHelper) Token(ctx context.Context, resp *AuthResponse) (string, bool, error) {
	a.lock.Lock()
	defer a.lock.Unlock()

	if resp == nil {
		// Token cannot be fetch if authentication is not valid
		return "", false, nil
	}

	v, ok := a.cache[resp.Username]
	if ok {
		// We received token
		if time.Now().Before(v.TTL) {
			return v.Token, true, nil
		}
		// Token has been expired
		delete(a.cache, resp.Username)
	}

	// We did not receive token, create one
	auth, err := a.client.CreateToken(ctx, &pbAuthenticationV1.CreateTokenRequest{
		Lifetime: durationpb.New(DefaultLifetime),
		User:     util.NewType(resp.Username),
	})
	if err != nil {
		return "", false, err
	}

	a.cache[resp.Username] = adbHelperToken{
		TTL:   time.Now().Add(DefaultTTL),
		Token: auth.GetToken(),
	}

	return auth.GetToken(), true, nil
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
			}, nil
		}
	}

	return nil, nil
}
