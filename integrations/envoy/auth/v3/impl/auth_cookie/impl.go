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
	goHttp "net/http"

	pbEnvoyAuthV3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"

	pbAuthenticationV1 "github.com/arangodb/kube-arangodb/integrations/authentication/v1/definition"
	pbImplEnvoyAuthV3Shared "github.com/arangodb/kube-arangodb/integrations/envoy/auth/v3/shared"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/cache"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

const JWTAuthorizationCookieName = "X-ArangoDB-Token-JWT"

func New(configuration pbImplEnvoyAuthV3Shared.Configuration) (pbImplEnvoyAuthV3Shared.AuthHandler, bool) {
	if !configuration.Extensions.JWT {
		return nil, false
	}

	var z impl

	z.configuration = configuration
	z.cache = cache.NewCache[pbImplEnvoyAuthV3Shared.Token, pbImplEnvoyAuthV3Shared.ResponseAuth](func(ctx context.Context, in pbImplEnvoyAuthV3Shared.Token) (pbImplEnvoyAuthV3Shared.ResponseAuth, error) {
		resp, err := z.configuration.AuthClient.Validate(ctx, &pbAuthenticationV1.ValidateRequest{
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

	return z, true
}

type impl struct {
	configuration pbImplEnvoyAuthV3Shared.Configuration

	cache cache.Cache[pbImplEnvoyAuthV3Shared.Token, pbImplEnvoyAuthV3Shared.ResponseAuth]
}

func (p impl) Handle(ctx context.Context, request *pbEnvoyAuthV3.CheckRequest, current *pbImplEnvoyAuthV3Shared.Response) error {
	if current.Authenticated() {
		// Already authenticated
		return nil
	}

	rawCookies := request.GetAttributes().GetRequest().GetHttp().GetHeaders()["cookie"]
	// Convert raw cookie string into map of http cookies
	header := goHttp.Header{}
	header.Add("Cookie", rawCookies)
	req := goHttp.Request{Header: header}
	cookies := req.Cookies()

	for _, cookie := range cookies {
		if cookie != nil {
			if cookie.Valid() != nil {
				continue
			}
			if cookie.Name == JWTAuthorizationCookieName {
				auth, err := p.cache.Get(ctx, pbImplEnvoyAuthV3Shared.Token(cookie.Value))
				if err != nil {
					logger.Err(err).Warn("Auth failure")
					return nil
				}

				current.User = &pbImplEnvoyAuthV3Shared.ResponseAuth{
					User:  auth.User,
					Roles: auth.Roles,
					Token: util.NewType(cookie.Value),
				}
			}

			current.Headers = append(current.Headers, pbImplEnvoyAuthV3Shared.FilterCookiesHeader(cookies, func(cookie *goHttp.Cookie) bool {
				return cookie.Valid() != nil
			}, func(cookie *goHttp.Cookie) bool {
				return cookie.Name == JWTAuthorizationCookieName
			})...)

			return nil
		}
	}

	return nil
}
