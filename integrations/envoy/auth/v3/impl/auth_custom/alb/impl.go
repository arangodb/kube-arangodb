//
// DISCLAIMER
//
// Copyright 2026 ArangoDB GmbH, Cologne, Germany
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

package alb

import (
	"context"
	"crypto/ecdsa"
	"time"

	pbEnvoyAuthV3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"

	pbImplEnvoyAuthV3Shared "github.com/arangodb/kube-arangodb/integrations/envoy/auth/v3/shared"
	platformAuthenticationApi "github.com/arangodb/kube-arangodb/pkg/apis/platform/v1beta1/authentication"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/cache"
)

func New(ctx context.Context, configuration pbImplEnvoyAuthV3Shared.Configuration) (pbImplEnvoyAuthV3Shared.AuthHandler, bool, error) {
	c := cache.NewConfigFile[platformAuthenticationApi.ALB](configuration.Auth.Path, time.Minute)

	i := &impl{
		fileConfig: c,
		resolver: cache.NewHashedConfiguration(c, func(ctx context.Context, in platformAuthenticationApi.ALB) (platformAuthenticationApi.ALBKeyResolver, error) {
			return in.KeyResolver()
		}),
	}

	// Cache the resolved signing keys by key id so a rotated set of ALB keys is picked up while a
	// steady state does not re-fetch on every request.
	i.keys = cache.NewCache(func(ctx context.Context, kid string) (*ecdsa.PublicKey, time.Time, error) {
		resolver, err := i.resolver.Get(ctx)
		if err != nil {
			return nil, util.Default[time.Time](), err
		}

		key, err := resolver(ctx, kid)
		if err != nil {
			return nil, util.Default[time.Time](), err
		}

		return key, time.Now().Add(time.Hour), nil
	})

	return i, true, nil
}

type impl struct {
	fileConfig cache.ConfigFile[platformAuthenticationApi.ALB]

	resolver cache.HashedConfiguration[platformAuthenticationApi.ALBKeyResolver]

	keys cache.Cache[string, *ecdsa.PublicKey]
}

func (i *impl) resolve(ctx context.Context, kid string) (*ecdsa.PublicKey, error) {
	return i.keys.Get(ctx, kid)
}

func (i *impl) Handle(ctx context.Context, request *pbEnvoyAuthV3.CheckRequest, current *pbImplEnvoyAuthV3Shared.Response) error {
	if current.Authenticated() {
		// Already authenticated
		return nil
	}

	raw, ok := request.GetAttributes().GetRequest().GetHttp().GetHeaders()[platformAuthenticationApi.ALBDataHeader]
	if !ok {
		return nil
	}

	cfg, _, err := i.fileConfig.Get(ctx)
	if err != nil {
		logger.Err(err).Warn("Unable to get ALB config")
		return nil
	}

	claims, err := cfg.VerifyToken(ctx, raw, i.resolve)
	if err != nil {
		logger.Err(err).Warn("ALB token verification failure")
		return nil
	}

	user, ok := claims[cfg.Claims.GetUsernameClaim()].(string)
	if !ok || user == "" {
		logger.Str("claim", cfg.Claims.GetUsernameClaim()).Warn("ALB token is missing the username claim")
		return nil
	}

	current.User = &pbImplEnvoyAuthV3Shared.ResponseAuth{
		User:   user,
		Groups: extractGroups(claims, cfg.Claims.GetGroupsClaim()),
	}

	return nil
}

// extractGroups reads the configured groups claim, accepting either a JSON array of strings or a
// single string value. An empty claim key or absent value yields no groups.
func extractGroups(claims map[string]interface{}, claim string) []string {
	if claim == "" {
		return nil
	}

	v, ok := claims[claim]
	if !ok {
		return nil
	}

	switch t := v.(type) {
	case string:
		if t == "" {
			return nil
		}
		return []string{t}
	case []interface{}:
		var groups []string
		for _, e := range t {
			if s, ok := e.(string); ok && s != "" {
				groups = append(groups, s)
			}
		}
		return groups
	default:
		return nil
	}
}
