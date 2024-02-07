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

package v1

import (
	"context"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/durationpb"

	pbAuthenticationV1 "github.com/arangodb/kube-arangodb/integrations/authentication/v1/definition"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/strings"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
	"github.com/arangodb/kube-arangodb/pkg/util/token"
)

func New(ctx context.Context, cfg Configuration) (svc.Handler, error) {
	return newInternal(ctx, cfg)
}

func newInternal(ctx context.Context, cfg Configuration) (*implementation, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	obj := &implementation{
		cfg: cfg,
		ctx: ctx,
	}

	return obj, nil
}

var _ pbAuthenticationV1.AuthenticationV1Server = &implementation{}
var _ svc.Handler = &implementation{}

type implementation struct {
	pbAuthenticationV1.UnimplementedAuthenticationV1Server

	lock sync.RWMutex

	ctx context.Context
	cfg Configuration

	cache *cache
}

func (i *implementation) Name() string {
	return "authentication.v1"
}

func (i *implementation) Health() svc.HealthState {
	return svc.Healthy
}

func (i *implementation) Register(registrar *grpc.Server) {
	pbAuthenticationV1.RegisterAuthenticationV1Server(registrar, i)
}

func (i *implementation) Validate(ctx context.Context, request *pbAuthenticationV1.ValidateRequest) (*pbAuthenticationV1.ValidateResponse, error) {
	if request == nil {
		return nil, errors.Errorf("Request is nil")
	}

	if !i.cfg.Enabled {
		return &pbAuthenticationV1.ValidateResponse{
			IsValid: true,
			Details: &pbAuthenticationV1.ValidateResponseDetails{
				Lifetime: durationpb.New(DefaultTokenMaxTTL),
				User:     DefaultAdminUser,
			},
		}, nil
	}

	cache, err := i.withCache()
	if err != nil {
		return nil, err
	}

	user, exp, err := i.extractTokenDetails(cache, request.GetToken())
	if err != nil {
		return &pbAuthenticationV1.ValidateResponse{
			IsValid: false,
			Message: err.Error(),
		}, nil
	}

	return &pbAuthenticationV1.ValidateResponse{
		IsValid: true,
		Details: &pbAuthenticationV1.ValidateResponseDetails{
			Lifetime: durationpb.New(exp),
			User:     user,
		},
	}, nil
}

func (i *implementation) CreateToken(ctx context.Context, request *pbAuthenticationV1.CreateTokenRequest) (*pbAuthenticationV1.CreateTokenResponse, error) {
	if request == nil {
		return nil, errors.Errorf("Request is nil")
	}

	if !i.cfg.Enabled {
		// Authentication is not enabled, pass with empty token

		return &pbAuthenticationV1.CreateTokenResponse{
			Lifetime: durationpb.New(DefaultTokenMaxTTL),
			User:     DefaultAdminUser,
			Token:    "",
		}, nil
	}

	cache, err := i.withCache()
	if err != nil {
		return nil, err
	}

	user := util.TypeOrDefault(request.User, i.cfg.Create.DefaultUser)
	duration := i.cfg.Create.DefaultTTL
	if v := request.Lifetime; v != nil {
		duration = v.AsDuration()
	}

	// Check configuration
	if v := i.cfg.Create.AllowedUsers; len(v) > 0 {
		if !strings.ListContains(v, user) {
			return nil, errors.Errorf("User %s is not allowed", user)
		}
	}

	if v := i.cfg.Create.MaxTTL; duration > v {
		duration = v
	}

	if v := i.cfg.Create.MinTTL; duration < v {
		duration = v
	}

	// Token is validated, we can continue with creation
	secret := cache.signingToken

	signedToken, err := token.New(secret, token.NewClaims().With(token.WithDefaultClaims(), token.WithCurrentIAT(), token.WithDuration(duration), token.WithUsername(user)))
	if err != nil {
		return nil, err
	}

	user, exp, err := i.extractTokenDetails(cache, signedToken)
	if err != nil {
		return nil, err
	}

	return &pbAuthenticationV1.CreateTokenResponse{
		Lifetime: durationpb.New(exp),
		User:     user,
		Token:    signedToken,
	}, nil
}

func (i *implementation) extractTokenDetails(cache *cache, t string) (string, time.Duration, error) {
	// Let's check if token is signed properly

	p, err := token.ParseWithAny(t, cache.validationTokens...)
	if err != nil {
		return "", 0, err
	}

	user := DefaultAdminUser
	if v, ok := p[token.ClaimPreferredUsername]; ok {
		if s, ok := v.(string); ok {
			user = s
		}
	}

	duration := DefaultTokenMaxTTL

	if v, ok := p[token.ClaimEXP]; ok {
		if s, ok := v.(int64); ok {
			duration = time.Until(time.Unix(s, 0))
		}
	}

	return user, duration, nil
}
