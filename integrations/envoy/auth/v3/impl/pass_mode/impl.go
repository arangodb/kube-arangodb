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

package pass_mode

import (
	"context"
	"fmt"
	goHttp "net/http"
	goStrings "strings"

	pbEnvoyCoreV3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	pbEnvoyAuthV3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	"google.golang.org/protobuf/types/known/durationpb"

	pbAuthenticationV1 "github.com/arangodb/kube-arangodb/integrations/authentication/v1/definition"
	pbImplEnvoyAuthV3Shared "github.com/arangodb/kube-arangodb/integrations/envoy/auth/v3/shared"
	networkingApi "github.com/arangodb/kube-arangodb/pkg/apis/networking/v1alpha1"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/cache"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func New(configuration pbImplEnvoyAuthV3Shared.Configuration) (pbImplEnvoyAuthV3Shared.AuthHandler, bool) {
	var z impl

	z.configuration = configuration
	z.cache = cache.NewHashCache[*pbImplEnvoyAuthV3Shared.ResponseAuth, pbImplEnvoyAuthV3Shared.Token](z.Token, pbImplEnvoyAuthV3Shared.DefaultTTL)

	return z, true
}

type impl struct {
	configuration pbImplEnvoyAuthV3Shared.Configuration

	cache cache.HashCache[*pbImplEnvoyAuthV3Shared.ResponseAuth, pbImplEnvoyAuthV3Shared.Token]
}

func (p impl) Handle(ctx context.Context, request *pbEnvoyAuthV3.CheckRequest, current *pbImplEnvoyAuthV3Shared.Response) error {
	if !current.Authenticated() {
		current.Headers = append(current.Headers,
			&pbEnvoyCoreV3.HeaderValueOption{
				Header: &pbEnvoyCoreV3.HeaderValue{
					Key:   pbImplEnvoyAuthV3Shared.AuthAuthenticatedHeader,
					Value: "false",
				},
				AppendAction: pbEnvoyCoreV3.HeaderValueOption_OVERWRITE_IF_EXISTS_OR_ADD,
			},
		)

		return nil
	}

	ext := request.GetAttributes().GetContextExtensions()

	switch networkingApi.ArangoRouteSpecAuthenticationPassMode(goStrings.ToLower(util.Optional(ext, pbImplEnvoyAuthV3Shared.AuthConfigAuthPassModeKey, ""))) {
	case networkingApi.ArangoRouteSpecAuthenticationPassModePass:
		if current.User.Token != nil {
			current.Headers = append(current.Headers, &pbEnvoyCoreV3.HeaderValueOption{
				Header: &pbEnvoyCoreV3.HeaderValue{
					Key:   pbImplEnvoyAuthV3Shared.AuthorizationHeader,
					Value: fmt.Sprintf("bearer %s", *current.User.Token),
				},
				AppendAction: pbEnvoyCoreV3.HeaderValueOption_OVERWRITE_IF_EXISTS_OR_ADD,
			},
			)
		} else {
			token, err := p.cache.Get(ctx, current.User)
			if err != nil {
				logger.Err(err).Error("Unable to get token")
				return pbImplEnvoyAuthV3Shared.DeniedResponse{
					Code: goHttp.StatusUnauthorized,
					Message: &pbImplEnvoyAuthV3Shared.DeniedMessage{
						Message: "Unable to render token",
					},
				}
			}

			current.Headers = append(current.Headers, &pbEnvoyCoreV3.HeaderValueOption{
				Header: &pbEnvoyCoreV3.HeaderValue{
					Key:   pbImplEnvoyAuthV3Shared.AuthorizationHeader,
					Value: fmt.Sprintf("bearer %s", token),
				},
				AppendAction: pbEnvoyCoreV3.HeaderValueOption_OVERWRITE_IF_EXISTS_OR_ADD,
			},
			)
		}
	case networkingApi.ArangoRouteSpecAuthenticationPassModeOverride:
		token, err := p.cache.Get(ctx, current.User)

		if err != nil {
			logger.Err(err).Error("Unable to get token")
			return pbImplEnvoyAuthV3Shared.DeniedResponse{
				Code: goHttp.StatusUnauthorized,
				Message: &pbImplEnvoyAuthV3Shared.DeniedMessage{
					Message: "Unable to render token",
				},
			}
		}

		current.Headers = append(current.Headers, &pbEnvoyCoreV3.HeaderValueOption{
			Header: &pbEnvoyCoreV3.HeaderValue{
				Key:   pbImplEnvoyAuthV3Shared.AuthorizationHeader,
				Value: fmt.Sprintf("bearer %s", token),
			},
			AppendAction: pbEnvoyCoreV3.HeaderValueOption_OVERWRITE_IF_EXISTS_OR_ADD,
		},
		)
	case networkingApi.ArangoRouteSpecAuthenticationPassModeRemove:
		current.Headers = append(current.Headers, &pbEnvoyCoreV3.HeaderValueOption{
			Header: &pbEnvoyCoreV3.HeaderValue{
				Key: pbImplEnvoyAuthV3Shared.AuthorizationHeader,
			},
			AppendAction:   pbEnvoyCoreV3.HeaderValueOption_OVERWRITE_IF_EXISTS_OR_ADD,
			KeepEmptyValue: false,
		},
		)
	}

	return nil
}

func (p impl) Token(ctx context.Context, in *pbImplEnvoyAuthV3Shared.ResponseAuth) (pbImplEnvoyAuthV3Shared.Token, error) {
	if in == nil {
		return "", errors.Errorf("Nil is not allowed")
	}
	resp, err := p.configuration.AuthClient.CreateToken(ctx, &pbAuthenticationV1.CreateTokenRequest{
		Lifetime: durationpb.New(pbImplEnvoyAuthV3Shared.DefaultLifetime),
		User:     util.NewType(in.User),
		Roles:    in.Roles,
	})
	if err != nil {
		return "", err
	}

	return pbImplEnvoyAuthV3Shared.Token(resp.Token), nil
}
