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

package v3

import (
	"context"
	goHttp "net/http"

	pbEnvoyAuthV3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
)

const JWTAuthorizationCookieName = "X-ArangoDB-Token-JWT"

func (i *impl) checkADBJWTCookie(ctx context.Context, request *pbEnvoyAuthV3.CheckRequest, current *AuthResponse) (*AuthResponse, error) {
	if current != nil {
		// Already authenticated
		return current, nil
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
				resp, err := i.helper.Validate(ctx, cookie.Value)
				if err != nil {
					logger.Err(err).Warn("Auth failure")
					return nil, nil
				}

				if resp == nil {
					return nil, nil
				}

				resp.Headers = filterCookiesHeader(cookies, func(cookie *goHttp.Cookie) bool {
					return cookie.Valid() != nil
				}, func(cookie *goHttp.Cookie) bool {
					return cookie.Name == JWTAuthorizationCookieName
				})

				return resp, nil
			}
		}
	}

	return nil, nil
}
