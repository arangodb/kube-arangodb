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
	"fmt"
	"net/http"
	"strings"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	pbEnvoyAuthV3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	"google.golang.org/grpc"

	pbAuthenticationV1 "github.com/arangodb/kube-arangodb/integrations/authentication/v1/definition"
	networkingApi "github.com/arangodb/kube-arangodb/pkg/apis/networking/v1alpha1"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/errors/panics"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
)

func New(authClient pbAuthenticationV1.AuthenticationV1Client) svc.Handler {
	return &impl{
		helper: NewADBHelper(authClient),
	}
}

var _ pbEnvoyAuthV3.AuthorizationServer = &impl{}
var _ svc.Handler = &impl{}

type impl struct {
	pbEnvoyAuthV3.UnimplementedAuthorizationServer

	helper ADBHelper
}

func (i *impl) Name() string {
	return Name
}

func (i *impl) Health() svc.HealthState {
	return svc.Healthy
}

func (i *impl) Register(registrar *grpc.Server) {
	pbEnvoyAuthV3.RegisterAuthorizationServer(registrar, i)
}

func (i *impl) Check(ctx context.Context, request *pbEnvoyAuthV3.CheckRequest) (*pbEnvoyAuthV3.CheckResponse, error) {
	resp, err := panics.RecoverO1(func() (*pbEnvoyAuthV3.CheckResponse, error) {
		return i.check(ctx, request)
	})

	if err != nil {
		var v DeniedResponse
		if errors.As(err, &v) {
			return v.GetCheckResponse()
		}
		return nil, err
	}
	return resp, nil
}

func (i *impl) check(ctx context.Context, request *pbEnvoyAuthV3.CheckRequest) (*pbEnvoyAuthV3.CheckResponse, error) {
	ext := request.GetAttributes().GetContextExtensions()

	if v, ok := ext[AuthConfigTypeKey]; !ok || v != AuthConfigTypeValue {
		return nil, DeniedResponse{
			Code: http.StatusBadRequest,
			Message: &DeniedMessage{
				Message: "Auth plugin is not enabled for this request",
			},
		}
	}

	authenticated, err := MergeAuthRequest(ctx, request, i.checkADBJWT)
	if err != nil {
		return nil, err
	}

	if util.Optional(ext, AuthConfigAuthRequiredKey, AuthConfigKeywordFalse) == AuthConfigKeywordTrue && authenticated == nil {
		return nil, DeniedResponse{
			Code: http.StatusUnauthorized,
			Message: &DeniedMessage{
				Message: "Unauthorized",
			},
		}
	}

	if authenticated != nil {
		var headers = []*corev3.HeaderValueOption{
			{
				Header: &corev3.HeaderValue{
					Key:   AuthUsernameHeader,
					Value: authenticated.Username,
				},
				AppendAction: corev3.HeaderValueOption_OVERWRITE_IF_EXISTS_OR_ADD,
			},
			{
				Header: &corev3.HeaderValue{
					Key:   AuthAuthenticatedHeader,
					Value: "true",
				},
				AppendAction: corev3.HeaderValueOption_OVERWRITE_IF_EXISTS_OR_ADD,
			},
		}

		switch networkingApi.ArangoRouteSpecAuthenticationPassMode(strings.ToLower(util.Optional(ext, AuthConfigAuthPassModeKey, ""))) {
		case networkingApi.ArangoRouteSpecAuthenticationPassModeOverride:
			token, ok, err := i.helper.Token(ctx, authenticated)
			if err != nil {
				return nil, err
			}

			if !ok {
				return nil, DeniedResponse{
					Code: http.StatusUnauthorized,
					Message: &DeniedMessage{
						Message: "Unable to render token",
					},
				}
			}

			headers = append(headers, &corev3.HeaderValueOption{
				Header: &corev3.HeaderValue{
					Key:   "authorization",
					Value: fmt.Sprintf("bearer %s", token),
				},
				AppendAction: corev3.HeaderValueOption_OVERWRITE_IF_EXISTS_OR_ADD,
			},
			)
		case networkingApi.ArangoRouteSpecAuthenticationPassModeRemove:
			headers = append(headers, &corev3.HeaderValueOption{
				Header: &corev3.HeaderValue{
					Key: "authorization",
				},
				AppendAction:   corev3.HeaderValueOption_OVERWRITE_IF_EXISTS_OR_ADD,
				KeepEmptyValue: false,
			},
			)
		}

		return &pbEnvoyAuthV3.CheckResponse{
			HttpResponse: &pbEnvoyAuthV3.CheckResponse_OkResponse{
				OkResponse: &pbEnvoyAuthV3.OkHttpResponse{
					Headers: headers,
				},
			},
		}, nil
	}

	return &pbEnvoyAuthV3.CheckResponse{
		HttpResponse: &pbEnvoyAuthV3.CheckResponse_OkResponse{
			OkResponse: &pbEnvoyAuthV3.OkHttpResponse{
				Headers: []*corev3.HeaderValueOption{
					{
						Header: &corev3.HeaderValue{
							Key:   AuthAuthenticatedHeader,
							Value: "false",
						},
						AppendAction: corev3.HeaderValueOption_OVERWRITE_IF_EXISTS_OR_ADD,
					},
				},
			},
		},
	}, nil
}
