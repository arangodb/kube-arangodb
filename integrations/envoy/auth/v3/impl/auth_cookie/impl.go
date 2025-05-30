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

package auth_cookie

import (
	"context"
	pbEnvoyCoreV3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	goHttp "net/http"
	goStrings "strings"

	pbEnvoyAuthV3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"

	pbAuthenticationV1 "github.com/arangodb/kube-arangodb/integrations/authentication/v1/definition"
	pbImplEnvoyAuthV3Shared "github.com/arangodb/kube-arangodb/integrations/envoy/auth/v3/shared"
	networkingApi "github.com/arangodb/kube-arangodb/pkg/apis/networking/v1alpha1"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/cache"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

const JWTAuthorizationCookieName = "X-ArangoDB-Token-JWT"

func New(configuration pbImplEnvoyAuthV3Shared.Configuration) (pbImplEnvoyAuthV3Shared.AuthHandler, bool) {
	if !configuration.Extensions.CookieJWT {
		logger.Info("Gateway CookieAuth Disabled")
		return nil, false
	}

	var z impl

	z.configuration = configuration
	z.authClient = cache.NewObject[pbAuthenticationV1.AuthenticationV1Client](configuration.GetAuthClientFetcher)
	z.cache = cache.NewCache[pbImplEnvoyAuthV3Shared.Token, pbImplEnvoyAuthV3Shared.ResponseAuth](func(ctx context.Context, in pbImplEnvoyAuthV3Shared.Token) (pbImplEnvoyAuthV3Shared.ResponseAuth, error) {
		client, err := z.authClient.Get(ctx)
		if err != nil {
			return pbImplEnvoyAuthV3Shared.ResponseAuth{}, err
		}

		resp, err := client.Validate(ctx, &pbAuthenticationV1.ValidateRequest{
			Token: string(in),
		})
		if err != nil {
			return pbImplEnvoyAuthV3Shared.ResponseAuth{}, err
		}

		if !resp.GetIsValid() {
			return pbImplEnvoyAuthV3Shared.ResponseAuth{}, errors.Errorf("Invalid Token: %s", resp.GetMessage())
		}

		if resp.Details == nil {
			return pbImplEnvoyAuthV3Shared.ResponseAuth{}, errors.Errorf("Missing Details: %s", resp.GetMessage())
		}

		return pbImplEnvoyAuthV3Shared.ResponseAuth{
			User:  resp.GetDetails().GetUser(),
			Roles: resp.GetDetails().GetRoles(),
		}, nil
	}, pbImplEnvoyAuthV3Shared.DefaultTTL)

	logger.Info("Gateway CookieAuth Enabled")
	return z, true
}

type impl struct {
	configuration pbImplEnvoyAuthV3Shared.Configuration

	cache cache.Cache[pbImplEnvoyAuthV3Shared.Token, pbImplEnvoyAuthV3Shared.ResponseAuth]

	authClient cache.Object[pbAuthenticationV1.AuthenticationV1Client]
}

func (p impl) Handle(ctx context.Context, request *pbEnvoyAuthV3.CheckRequest, current *pbImplEnvoyAuthV3Shared.Response) error {
	if current.Authenticated() {
		// Already authenticated
		return nil
	}

	for _, cookie := range pbImplEnvoyAuthV3Shared.ExtractRequestCookies(request).Filter(func(in *goHttp.Cookie) bool {
		return in.Name == JWTAuthorizationCookieName
	}).Get() {
		auth, err := p.cache.Get(ctx, pbImplEnvoyAuthV3Shared.Token(cookie.Value))
		if err != nil {
			logger.Err(err).Warn("Auth failure")
			continue
		}

		current.User = &pbImplEnvoyAuthV3Shared.ResponseAuth{
			User:  auth.User,
			Roles: auth.Roles,
			Token: util.NewType(cookie.Value),
		}

		ext := request.GetAttributes().GetContextExtensions()

		switch networkingApi.ArangoRouteSpecAuthenticationPassMode(goStrings.ToLower(util.Optional(ext, pbImplEnvoyAuthV3Shared.AuthConfigAuthPassModeKey, ""))) {
		case networkingApi.ArangoRouteSpecAuthenticationPassModePass:
			// Keep headers
		default:
			current.Headers = []*pbEnvoyCoreV3.HeaderValueOption{
				{
					Header: &pbEnvoyCoreV3.HeaderValue{
						Key: pbImplEnvoyAuthV3Shared.CookieHeader,
					},
					AppendAction:   pbEnvoyCoreV3.HeaderValueOption_OVERWRITE_IF_EXISTS,
					KeepEmptyValue: false,
				},
			}
		}
		return nil
	}

	return nil
}
